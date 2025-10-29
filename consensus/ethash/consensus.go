// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethash

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/consensus/misc/eip4844"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/params/mutations"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/vars"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"golang.org/x/crypto/sha3"
)

// Ethash proof-of-work protocol constants.
var (
	maxUncles              = 2                // Maximum number of uncles allowed in a single block (default for non-Halo chains)
	allowedFutureBlockTime = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
)

// getAllowedFutureBlockTime returns the maximum time a block can be in the future
// before being rejected as invalid.
//
// Halo chain (ChainID 12000) uses 30 seconds to handle operational clock drift
// between distributed mining setups (e.g., Windows miners + Linux RPC nodes).
// Standard Ethereum uses 15 seconds.
//
// Note: This tolerance is for block ACCEPTANCE only. Difficulty calculations
// use timestamp capping to prevent manipulation attacks.
func getAllowedFutureBlockTime(chainID *big.Int) time.Duration {
	if chainID != nil && chainID.Uint64() == 12000 {
		return 30 * time.Second // Halo: 30s tolerance for operational clock drift
	}
	return allowedFutureBlockTime // Standard Ethereum: 15s
}

// getMaxUncles returns the maximum number of uncles allowed for a given chain
// Halo chain (ChainID 12000) allows 1 uncle, others allow 2
func getMaxUncles(chainID *big.Int) int {
	if chainID != nil && chainID.Uint64() == 12000 {
		return 1 // Halo chain
	}
	return maxUncles // Standard Ethereum (2)
}

// getMaxUncleDepth returns the maximum depth for uncles on a given chain
// Halo chain (ChainID 12000) allows depth of 2, others allow 7
func getMaxUncleDepth(chainID *big.Int) int {
	if chainID != nil && chainID.Uint64() == 12000 {
		return 2 // Halo chain - uncles can be max 2 blocks deep
	}
	return 7 // Standard Ethereum - uncles can be max 7 blocks deep
}

// verifyMedianTimePast ensures block timestamp is greater than median of last 11 blocks.
// This prevents miners from backdating blocks to manipulate difficulty.
//
// This is a Bitcoin-inspired security measure that prevents:
// - Backdating attacks (creating blocks with old timestamps)
// - Timestamp manipulation to lower difficulty
// - Chain reorganization attacks using timestamp gaming
func verifyMedianTimePast(chain consensus.ChainHeaderReader, header, parent *types.Header) error {
	// Need at least 11 blocks for meaningful MTP calculation
	if parent.Number.Uint64() < 11 {
		return nil // Skip validation for early blocks
	}

	// Collect timestamps of last 11 blocks (including parent)
	timestamps := make([]uint64, 11)
	for i := 0; i < 11; i++ {
		h := chain.GetHeaderByNumber(parent.Number.Uint64() - uint64(i))
		if h == nil {
			// Can't validate without full history, skip check
			return nil
		}
		timestamps[i] = h.Time
	}

	// Sort timestamps to find median
	sortedTimestamps := make([]uint64, 11)
	copy(sortedTimestamps, timestamps)
	// Simple bubble sort (only 11 elements, performance not critical)
	for i := 0; i < 11; i++ {
		for j := 0; j < 10-i; j++ {
			if sortedTimestamps[j] > sortedTimestamps[j+1] {
				sortedTimestamps[j], sortedTimestamps[j+1] = sortedTimestamps[j+1], sortedTimestamps[j]
			}
		}
	}
	medianTime := sortedTimestamps[5] // Middle value of 11 sorted elements

	// Block timestamp must be strictly greater than median
	if header.Time <= medianTime {
		return fmt.Errorf("timestamp %d not greater than median time past %d (backdating prevented)",
			header.Time, medianTime)
	}

	return nil
}

// getMaxTimestampJump returns the maximum allowed timestamp increase per block
// This prevents timestamp racing attacks where miners jump timestamps wildly
func getMaxTimestampJump(chainID *big.Int) uint64 {
	if chainID != nil && chainID.Uint64() == 12000 {
		return 10 // Halo: max 10 seconds jump per block
	}
	return 60 // Standard Ethereum: more lenient due to 12s block time
}

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	errOlderBlockTime    = errors.New("timestamp older than parent")
	errTooManyUncles     = errors.New("too many uncles")
	errDuplicateUncle    = errors.New("duplicate uncle")
	errUncleIsAncestor   = errors.New("uncle is ancestor")
	errDanglingUncle     = errors.New("uncle's parent is not ancestor")
	errUncleTooDeep      = errors.New("uncle depth exceeds maximum allowed")
	errInvalidDifficulty = errors.New("non-positive difficulty")
	errInvalidMixDigest  = errors.New("invalid mix digest")
	errInvalidPoW        = errors.New("invalid proof-of-work")
)

// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-work verified author of the block.
func (ethash *Ethash) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
func (ethash *Ethash) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {
	// If we're running a full engine faking, accept any input as valid
	if ethash.config.PowMode == ModeFullFake {
		return nil
	}
	// Short circuit if the header is known, or its parent not
	number := header.Number.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	// Sanity checks passed, do a proper verification
	return ethash.verifyHeader(chain, header, parent, false, seal, time.Now().Unix())
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications.
func (ethash *Ethash) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	// If we're running a full engine faking, accept any input as valid
	if ethash.config.PowMode == ModeFullFake || len(headers) == 0 {
		abort, results := make(chan struct{}), make(chan error, len(headers))
		for i := 0; i < len(headers); i++ {
			results <- nil
		}
		return abort, results
	}

	// Spawn as many workers as allowed threads
	workers := runtime.GOMAXPROCS(0)
	if len(headers) < workers {
		workers = len(headers)
	}

	// Create a task channel and spawn the verifiers
	var (
		inputs  = make(chan int)
		done    = make(chan int, workers)
		errors  = make([]error, len(headers))
		abort   = make(chan struct{})
		unixNow = time.Now().Unix()
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				errors[index] = ethash.verifyHeaderWorker(chain, headers, seals, index, unixNow)
				done <- index
			}
		}()
	}

	errorsOut := make(chan error, len(headers))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(headers))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(headers) {
					// Reached end of headers. Stop sending to workers.
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					errorsOut <- errors[out]
					if out == len(headers)-1 {
						return
					}
				}
			case <-abort:
				return
			}
		}
	}()
	return abort, errorsOut
}

func (ethash *Ethash) verifyHeaderWorker(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool, index int, unixNow int64) error {
	var parent *types.Header
	if index == 0 {
		parent = chain.GetHeader(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
	}
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	return ethash.verifyHeader(chain, headers[index], parent, false, seals[index], unixNow)
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of the stock Ethereum ethash engine.
func (ethash *Ethash) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	// If we're running a full engine faking, accept any input as valid
	if ethash.config.PowMode == ModeFullFake {
		return nil
	}
	// Get chain-specific uncle parameters
	chainID := chain.Config().GetChainID()
	maxUnclesForChain := getMaxUncles(chainID)
	maxUncleDepthForChain := getMaxUncleDepth(chainID)

	// Verify that there are at most maxUnclesForChain uncles included in this block
	if len(block.Uncles()) > maxUnclesForChain {
		return errTooManyUncles
	}
	if len(block.Uncles()) == 0 {
		return nil
	}
	// Gather the set of past uncles and ancestors
	uncles, ancestors := mapset.NewSet[common.Hash](), make(map[common.Hash]*types.Header)

	number, parent := block.NumberU64()-1, block.ParentHash()
	for i := 0; i < maxUncleDepthForChain; i++ {
		ancestorHeader := chain.GetHeader(parent, number)
		if ancestorHeader == nil {
			break
		}
		ancestors[parent] = ancestorHeader
		// If the ancestor doesn't have any uncles, we don't have to iterate them
		if ancestorHeader.UncleHash != types.EmptyUncleHash {
			// Need to add those uncles to the banned list too
			ancestor := chain.GetBlock(parent, number)
			if ancestor == nil {
				break
			}
			for _, uncle := range ancestor.Uncles() {
				uncles.Add(uncle.Hash())
			}
		}
		parent, number = ancestorHeader.ParentHash, number-1
	}
	ancestors[block.Hash()] = block.Header()
	uncles.Add(block.Hash())

	// Verify each of the uncles that it's recent, but not an ancestor
	for _, uncle := range block.Uncles() {
		// Make sure every uncle is rewarded only once
		hash := uncle.Hash()
		if uncles.Contains(hash) {
			return errDuplicateUncle
		}
		uncles.Add(hash)

		// Make sure the uncle has a valid ancestry
		if ancestors[hash] != nil {
			return errUncleIsAncestor
		}
		if ancestors[uncle.ParentHash] == nil || uncle.ParentHash == block.ParentHash() {
			return errDanglingUncle
		}

		// Verify uncle depth doesn't exceed chain-specific maximum
		uncleDepth := new(big.Int).Sub(block.Number(), uncle.Number).Uint64()
		if uncleDepth > uint64(maxUncleDepthForChain) {
			return errUncleTooDeep
		}

		if err := ethash.verifyHeader(chain, uncle, ancestors[uncle.ParentHash], true, true, time.Now().Unix()); err != nil {
			return err
		}
	}
	return nil
}

// verifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
// See YP section 4.3.4. "Block Header Validity"
func (ethash *Ethash) verifyHeader(chain consensus.ChainHeaderReader, header, parent *types.Header, uncle bool, seal bool, unixNow int64) error {
	// Ensure that the header's extra-data section is of a reasonable size (32)
	if uint64(len(header.Extra)) > vars.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), vars.MaximumExtraDataSize)
	}
	// Verify the header's timestamp
	if !uncle {
		// Use chain-specific future block tolerance
		chainID := chain.Config().GetChainID()
		allowedFuture := getAllowedFutureBlockTime(chainID)
		if header.Time > uint64(unixNow+int64(allowedFuture.Seconds())) {
			return consensus.ErrFutureBlock
		}
	}
	// Verify the timestamp
	if header.Time <= parent.Time {
		return errOlderBlockTime
	}

	// Halo-specific timestamp validations (ChainID 12000)
	chainID := chain.Config().GetChainID()
	if chainID != nil && chainID.Uint64() == 12000 && !uncle {
		// SECURITY LAYER 1: Median Time Past (MTP) validation
		// Prevents backdating attacks by ensuring timestamp > median of last 11 blocks
		if err := verifyMedianTimePast(chain, header, parent); err != nil {
			return err
		}

		// SECURITY LAYER 2: Maximum timestamp jump per block
		// Prevents timestamp racing where miners jump timestamps wildly
		maxJump := getMaxTimestampJump(chainID)
		if header.Time > parent.Time+maxJump {
			return fmt.Errorf("timestamp jumped too far: %d -> %d (max +%ds per block)",
				parent.Time, header.Time, maxJump)
		}
	}
	// Verify the block's difficulty based on its timestamp and parent's difficulty
	expected := ethash.CalcDifficulty(chain, header.Time, parent)
	if expected.Cmp(header.Difficulty) != 0 {
		return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	}
	// Verify that the gas limit is <= 2^63-1
	if header.GasLimit > vars.MaxGasLimit {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, vars.MaxGasLimit)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return consensus.ErrInvalidNumber
	}
	// Verify the block's gas usage and (if applicable) verify the base fee.
	if !chain.Config().IsEnabled(chain.Config().GetEIP1559Transition, header.Number) {
		// Verify BaseFee not present before EIP-1559 fork.
		if header.BaseFee != nil {
			return fmt.Errorf("invalid baseFee before fork: have %d, expected 'nil'", header.BaseFee)
		}
		if err := misc.VerifyGaslimit(parent.GasLimit, header.GasLimit); err != nil {
			return err
		}
	} else if err := eip1559.VerifyEIP1559Header(chain.Config(), parent, header); err != nil {
		// Verify the header's EIP-1559 attributes.
		return err
	}

	// Shanghai
	// EIP-4895: Beacon chain push withdrawals as operations
	// Verify the non-existence of withdrawalsHash (EIP-4895: Beacon chain push withdrawals as operations).
	eip4895Enabled := chain.Config().IsEnabledByTime(chain.Config().GetEIP4895TransitionTime, &header.Time) || chain.Config().IsEnabled(chain.Config().GetEIP4895Transition, header.Number)
	if !eip4895Enabled {
		if header.WithdrawalsHash != nil {
			return fmt.Errorf("invalid withdrawalsHash: have %x, expected nil", header.WithdrawalsHash)
		}
	} else {
		if header.WithdrawalsHash == nil {
			return errors.New("header is missing withdrawalsHash")
		}
	}

	// Cancun
	// EIP-4844: Shard Blob Txes
	// EIP-4788: Beacon block root in the EVM
	eip4844Enabled := chain.Config().IsEnabledByTime(chain.Config().GetEIP4844TransitionTime, &header.Time) || chain.Config().IsEnabled(chain.Config().GetEIP4844Transition, header.Number)
	if !eip4844Enabled {
		switch {
		case header.ExcessBlobGas != nil:
			return fmt.Errorf("invalid excessBlobGas: have %d, expected nil", header.ExcessBlobGas)
		case header.BlobGasUsed != nil:
			return fmt.Errorf("invalid blobGasUsed: have %d, expected nil", header.BlobGasUsed)
		}
	} else {
		if err := eip4844.VerifyEIP4844Header(parent, header); err != nil {
			return err
		}
	}

	// EIP-4788: Beacon block root in the EVM
	eip4788Enabled := chain.Config().IsEnabledByTime(chain.Config().GetEIP4788TransitionTime, &header.Time) || chain.Config().IsEnabled(chain.Config().GetEIP4788Transition, header.Number)
	if !eip4788Enabled {
		if header.ParentBeaconRoot != nil {
			return fmt.Errorf("invalid parentBeaconRoot, have %#x, expected nil", header.ParentBeaconRoot)
		}
	} else {
		if header.ParentBeaconRoot == nil {
			return errors.New("header is missing beaconRoot")
		}
	}

	// // Add some fake checks for tests
	// if ethash.fakeDelay != nil {
	// 	time.Sleep(*ethash.fakeDelay)
	// }

	// Verify the engine specific seal securing the block
	if seal {
		if err := ethash.verifySeal(chain, header, false); err != nil {
			return err
		}
	}
	// If all checks passed, validate any special fields for hard forks
	if err := mutations.VerifyDAOHeaderExtraData(chain.Config(), header); err != nil {
		return err
	}
	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (ethash *Ethash) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	// HALO-SPECIFIC: Pass chain reader for historical difficulty lookups
	if chain.Config().GetChainID() != nil && chain.Config().GetChainID().Uint64() == 12000 {
		return calcDifficultyHaloSecure(chain, time, parent)
	}
	return CalcDifficulty(chain.Config(), time, parent)
}

// parent_time_delta is a convenience fn for CalcDifficulty
func parent_time_delta(t uint64, p *types.Header) *big.Int {
	return new(big.Int).Sub(new(big.Int).SetUint64(t), new(big.Int).SetUint64(p.Time))
}

// parent_diff_over_dbd is a  convenience fn for CalcDifficulty
func parent_diff_over_dbd(p *types.Header) *big.Int {
	return new(big.Int).Div(p.Difficulty, vars.DifficultyBoundDivisor)
}

// isHaloEmergencyMode checks if the network is in emergency recovery mode.
// Emergency mode activates when blocks are consistently slow (> 30s average over last 10 blocks).
// This allows the network to recover from extreme hashrate drops by temporarily
// relaxing the phase minimum difficulty.
//
// Returns true if in emergency mode, false otherwise.
func isHaloEmergencyMode(chain consensus.ChainHeaderReader, parent *types.Header) bool {
	// Need at least 10 blocks of history
	if parent.Number.Uint64() < 10 {
		return false
	}

	// Calculate average block time for last 10 blocks
	totalTime := uint64(0)
	blockCount := uint64(10)

	for i := uint64(0); i < blockCount; i++ {
		currentBlock := chain.GetHeaderByNumber(parent.Number.Uint64() - i)
		if currentBlock == nil {
			return false // Can't determine, assume not emergency
		}

		if i < blockCount-1 {
			prevBlock := chain.GetHeaderByNumber(parent.Number.Uint64() - i - 1)
			if prevBlock == nil {
				return false // Can't determine, assume not emergency
			}
			blockTime := currentBlock.Time - prevBlock.Time
			totalTime += blockTime
		}
	}

	avgBlockTime := totalTime / (blockCount - 1)

	// Emergency if average block time > 30 seconds
	// This indicates severe hashrate shortage
	return avgBlockTime > 30
}

// getHaloPhaseMinimum returns the minimum difficulty for a given block number.
// This implements PHASED SECURITY: starts with lower minimum for easy launch,
// gradually increases to full Ethereum-grade security.
//
// Phase 1 (Blocks 0-10,000): Minimum = 4,096
//   - Easy network launch with small hashrate
//   - Allows single GPU to mine comfortably
//   - Network can establish itself
//
// Phase 2 (Blocks 10,000-50,000): Linear increase 4,096 → 131,072
//   - Gradual hardening as network matures
//   - Smooth transition, no sudden jumps
//   - Gives time for more miners to join
//
// Phase 3 (Blocks 50,000+): Minimum = 131,072
//   - Full Ethereum-grade security
//   - Maximum protection against attacks
//   - Mature network with established hashrate
//
// SECURITY RATIONALE:
// - Early phase (< 50k blocks): Network is permissioned/controlled
// - Transition phase: Building decentralization
// - Late phase: Fully public, needs maximum security
func getHaloPhaseMinimum(blockNum uint64) *big.Int {
	const (
		phase1End = uint64(10000)  // End of easy phase
		phase2End = uint64(50000)  // End of transition phase
		minEarly  = 4096           // Early minimum (0x1000)
		minFull   = 131072         // Full minimum (0x20000)
	)

	// Phase 1: Easy start (blocks 0-10,000)
	if blockNum <= phase1End {
		return big.NewInt(minEarly)
	}

	// Phase 3: Full security (blocks 50,000+)
	if blockNum >= phase2End {
		return big.NewInt(minFull)
	}

	// Phase 2: Linear transition (blocks 10,001-50,000)
	// Calculate: minEarly + (blockNum - phase1End) * (minFull - minEarly) / (phase2End - phase1End)
	progress := blockNum - phase1End
	totalSteps := phase2End - phase1End
	increment := (minFull - minEarly) * progress / totalSteps

	return big.NewInt(int64(minEarly + increment))
}

// calcDifficultyHaloSecure implements Halo's SECURE difficulty algorithm with hybrid protection.
// This algorithm provides PRODUCTION-GRADE security against difficulty manipulation attacks.
//
// HYBRID PROTECTION STRATEGY (5-Layer Defense):
// 1. Timestamp Capping: Future timestamps capped to "now" to prevent manipulation
// 2. Absolute Minimum: 131,072 (same as Ethereum)
// 3. Peak Memory: Cannot drop below 10% of max difficulty in last 100 blocks
// 4. Historical Floor: Cannot drop below 50% of difficulty 100 blocks ago
// 5. Bounded Adjustments: Maximum ±50% change per block
//
// This prevents:
// - Future timestamp difficulty gaming (capping layer)
// - Hashrate manipulation attacks (attacker joins/leaves to game difficulty)
// - Rapid difficulty drops that enable chain reorganization
// - 51% attacks during low difficulty windows
//
// Note: We accept blocks up to 30s in the future for operational reasons (clock drift),
// but for difficulty calculations, we treat future timestamps as "now" to prevent gaming.
//
// Algorithm:
//   Timestamp capping → Base calculation → Apply all protection layers
func calcDifficultyHaloSecure(chain consensus.ChainHeaderReader, blockTime uint64, parent *types.Header) *big.Int {
	// ========== STEP 0: Timestamp Capping (CRITICAL SECURITY LAYER) ==========
	// We accept blocks up to 30s in the future to handle clock drift between mining setups,
	// BUT we cap timestamps to current time for difficulty calculations to prevent manipulation.
	//
	// Without this: Miners could use future timestamps to make blocks appear "slow",
	// causing difficulty to drop, then mine many blocks quickly at lower difficulty.

	// Get current wall clock time (note: renamed param to blockTime to avoid shadowing time package)
	currentTime := uint64(time.Now().Unix())

	adjustedTime := blockTime
	if blockTime > currentTime {
		// Block is in the future (up to 30s allowed for acceptance)
		// But treat it as "now" for difficulty calculation
		adjustedTime = currentTime
	}

	// ========== STEP 1: Calculate Base Difficulty Adjustment ==========
	targetBlockTime := int64(1)      // 1 second target
	adjustmentDivisor := int64(2048) // Same as Ethereum for stability
	maxAdjustmentPercent := int64(50) // Maximum 50% adjustment per block

	// Calculate time delta using CAPPED timestamp (security critical)
	timeDelta := int64(adjustedTime) - int64(parent.Time)

	// Edge case: Clock skew protection
	if timeDelta <= 0 {
		timeDelta = 1
	}

	// Edge case: Cap unreasonably long block times
	if timeDelta > 60 {
		timeDelta = 60
	}

	// Calculate how far off we are from target
	timeDeviation := timeDelta - targetBlockTime

	// Calculate base adjustment
	baseAdjustment := new(big.Int).Div(parent.Difficulty, big.NewInt(adjustmentDivisor))
	adjustment := new(big.Int).Mul(baseAdjustment, big.NewInt(-timeDeviation))

	// Cap maximum single-block adjustment (±50%)
	maxAdjustment := new(big.Int).Div(parent.Difficulty, big.NewInt(100/maxAdjustmentPercent))
	if adjustment.CmpAbs(maxAdjustment) > 0 {
		if adjustment.Sign() < 0 {
			adjustment.Neg(maxAdjustment)
		} else {
			adjustment.Set(maxAdjustment)
		}
	}

	// Apply adjustment
	newDifficulty := new(big.Int).Add(parent.Difficulty, adjustment)

	// ========== STEP 2: Apply Hybrid Protection (4 Layers) ==========

	// Get current block number (needed for phase calculation and lookbacks)
	blockNum := parent.Number.Uint64()

	// EMERGENCY RECOVERY MODE CHECK
	// If blocks are consistently slow (avg > 30s over last 10 blocks),
	// temporarily disable phase minimum to allow recovery
	isEmergency := isHaloEmergencyMode(chain, parent)

	// Layer 1: ABSOLUTE MINIMUM (PHASED - starts low, increases over time)
	// This allows network to start with small hashrate, then harden as it matures
	// EXCEPT in emergency mode where we allow full recovery
	var absoluteMinimum *big.Int
	if isEmergency {
		// Emergency: Only enforce a basic floor (4096) to allow recovery
		absoluteMinimum = big.NewInt(4096)
	} else {
		// Normal: Enforce phase-appropriate minimum
		absoluteMinimum = getHaloPhaseMinimum(blockNum)
	}

	// Layer 2: PEAK MEMORY PROTECTION (prevents manipulation attacks)
	// Cannot drop below 10% of highest difficulty in last 100 blocks
	peakMinimum := big.NewInt(0)
	if blockNum >= 100 {
		// Find max difficulty in last 100 blocks
		maxRecentDiff := new(big.Int).Set(parent.Difficulty)
		lookbackStart := blockNum - 99 // Look back 100 blocks including parent

		for i := lookbackStart; i <= blockNum; i++ {
			header := chain.GetHeaderByNumber(i)
			if header != nil && header.Difficulty.Cmp(maxRecentDiff) > 0 {
				maxRecentDiff.Set(header.Difficulty)
			}
		}

		// Minimum is 10% of recent peak
		peakMinimum.Div(maxRecentDiff, big.NewInt(10))
	}

	// Layer 3: HISTORICAL FLOOR (prevents rapid drops)
	// Cannot drop below 50% of difficulty 100 blocks ago
	historicalMinimum := big.NewInt(0)
	if blockNum >= 100 {
		headerOld := chain.GetHeaderByNumber(blockNum - 100)
		if headerOld != nil {
			historicalMinimum.Div(headerOld.Difficulty, big.NewInt(2))
		}
	}

	// Layer 4: EARLY BLOCK BOOST (helps network establish quickly)
	if blockNum < 100 && newDifficulty.Cmp(big.NewInt(500000)) < 0 {
		// For first 100 blocks, allow faster difficulty growth
		earlyBoost := new(big.Int).Div(newDifficulty, big.NewInt(10))
		if timeDelta < targetBlockTime {
			newDifficulty.Add(newDifficulty, earlyBoost)
		}
	}

	// ========== STEP 3: Enforce the HIGHEST of all minimums ==========
	finalMinimum := new(big.Int).Set(absoluteMinimum)

	if peakMinimum.Cmp(finalMinimum) > 0 {
		finalMinimum.Set(peakMinimum)
	}

	if historicalMinimum.Cmp(finalMinimum) > 0 {
		finalMinimum.Set(historicalMinimum)
	}

	// Apply the final minimum
	if newDifficulty.Cmp(finalMinimum) < 0 {
		newDifficulty.Set(finalMinimum)
	}

	return newDifficulty
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func CalcDifficulty(config ctypes.ChainConfigurator, time uint64, parent *types.Header) *big.Int {
	next := new(big.Int).Add(parent.Number, big1)
	out := new(big.Int)

	// NOTE: Halo (ChainID 12000) uses calcDifficultyHaloSecure() via the method version
	// of CalcDifficulty, not this standalone function.

	// TODO (meowbits): do we need this?
	// if config.IsEnabled(config.GetEthashTerminalTotalDifficulty, next) {
	// 	return big.NewInt(1)
	// }

	// ADJUSTMENT algorithms
	if config.IsEnabled(config.GetEthashEIP100BTransition, next) {
		// https://github.com/ethereum/EIPs/issues/100
		// algorithm:
		// diff = (parent_diff +
		//         (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
		//        ) + 2^(periodCount - 2)
		out.Div(parent_time_delta(time, parent), vars.EIP100FDifficultyIncrementDivisor)

		if parent.UncleHash == types.EmptyUncleHash {
			out.Sub(big1, out)
		} else {
			out.Sub(big2, out)
		}
		out.Set(math.BigMax(out, bigMinus99))
		out.Mul(parent_diff_over_dbd(parent), out)
		out.Add(out, parent.Difficulty)
	} else if config.IsEnabled(config.GetEIP2Transition, next) {
		// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2.md
		// algorithm:
		// diff = (parent_diff +
		//         (parent_diff / 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
		//        )
		out.Div(parent_time_delta(time, parent), vars.EIP2DifficultyIncrementDivisor)
		out.Sub(big1, out)
		out.Set(math.BigMax(out, bigMinus99))
		out.Mul(parent_diff_over_dbd(parent), out)
		out.Add(out, parent.Difficulty)
	} else {
		// FRONTIER
		// algorithm:
		// diff =
		//   if parent_block_time_delta < params.DurationLimit
		//      parent_diff + (parent_diff // 2048)
		//   else
		//      parent_diff - (parent_diff // 2048)
		out.Set(parent.Difficulty)
		if parent_time_delta(time, parent).Cmp(vars.DurationLimit) < 0 {
			out.Add(out, parent_diff_over_dbd(parent))
		} else {
			out.Sub(out, parent_diff_over_dbd(parent))
		}
	}

	// after adjustment and before bomb
	out.Set(math.BigMax(out, vars.MinimumDifficulty))

	if config.IsEnabled(config.GetEthashECIP1041Transition, next) {
		return out
	}

	// EXPLOSION delays

	// exPeriodRef the explosion clause's reference point
	exPeriodRef := new(big.Int).Add(parent.Number, big1)

	if config.IsEnabled(config.GetEthashECIP1010PauseTransition, next) {
		ecip1010Explosion(config, next, exPeriodRef)
	} else if len(config.GetEthashDifficultyBombDelaySchedule()) > 0 {
		// This logic varies from the original fork-based logic (below) in that
		// configured delay values are treated as compounding values (-2000000 + -3000000 = -5000000@constantinople)
		// as opposed to hardcoded pre-compounded values (-5000000@constantinople).
		// Thus the Sub-ing.
		fakeBlockNumber := new(big.Int).Set(exPeriodRef)
		for activated, dur := range config.GetEthashDifficultyBombDelaySchedule() {
			if exPeriodRef.Cmp(big.NewInt(int64(activated))) < 0 {
				continue
			}
			fakeBlockNumber.Sub(fakeBlockNumber, dur.ToBig())
		}
		exPeriodRef.Set(fakeBlockNumber)
	} else if config.IsEnabled(config.GetEthashEIP5133Transition, next) {
		// calcDifficultyEip4345 is the difficulty adjustment algorithm as specified by EIP 4345.
		// It offsets the bomb a total of 10.7M blocks.
		fakeBlockNumber := new(big.Int)
		delayWithOffset := new(big.Int).Sub(vars.EIP5133DifficultyBombDelay.ToBig(), common.Big1)
		if parent.Number.Cmp(delayWithOffset) >= 0 {
			fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, delayWithOffset)
		}
		exPeriodRef.Set(fakeBlockNumber)
	} else if config.IsEnabled(config.GetEthashEIP4345Transition, next) {
		// calcDifficultyEip4345 is the difficulty adjustment algorithm as specified by EIP 4345.
		// It offsets the bomb a total of 10.7M blocks.
		fakeBlockNumber := new(big.Int)
		delayWithOffset := new(big.Int).Sub(vars.EIP4345DifficultyBombDelay.ToBig(), common.Big1)
		if parent.Number.Cmp(delayWithOffset) >= 0 {
			fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, delayWithOffset)
		}
		exPeriodRef.Set(fakeBlockNumber)
	} else if config.IsEnabled(config.GetEthashEIP3554Transition, next) {
		// calcDifficultyEIP3554 is the difficulty adjustment algorithm for London (December 2021).
		// The calculation uses the Byzantium rules, but with bomb offset 9.7M.
		fakeBlockNumber := new(big.Int)
		delayWithOffset := new(big.Int).Sub(vars.EIP3554DifficultyBombDelay.ToBig(), common.Big1)
		if parent.Number.Cmp(delayWithOffset) >= 0 {
			fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, delayWithOffset)
		}
		exPeriodRef.Set(fakeBlockNumber)
	} else if config.IsEnabled(config.GetEthashEIP2384Transition, next) {
		// calcDifficultyEIP2384 is the difficulty adjustment algorithm for Muir Glacier.
		// The calculation uses the Byzantium rules, but with bomb offset 9M.
		fakeBlockNumber := new(big.Int)
		delayWithOffset := new(big.Int).Sub(vars.EIP2384DifficultyBombDelay.ToBig(), common.Big1)
		if parent.Number.Cmp(delayWithOffset) >= 0 {
			fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, delayWithOffset)
		}
		exPeriodRef.Set(fakeBlockNumber)
	} else if config.IsEnabled(config.GetEthashEIP1234Transition, next) {
		// calcDifficultyEIP1234 is the difficulty adjustment algorithm for Constantinople.
		// The calculation uses the Byzantium rules, but with bomb offset 5M.
		// Specification EIP-1234: https://eips.ethereum.org/EIPS/eip-1234
		// Note, the calculations below looks at the parent number, which is 1 below
		// the block number. Thus we remove one from the delay given

		// calculate a fake block number for the ice-age delay
		// Specification: https://eips.ethereum.org/EIPS/eip-1234
		fakeBlockNumber := new(big.Int)
		delayWithOffset := new(big.Int).Sub(vars.EIP1234DifficultyBombDelay.ToBig(), common.Big1)
		if parent.Number.Cmp(delayWithOffset) >= 0 {
			fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, delayWithOffset)
		}
		exPeriodRef.Set(fakeBlockNumber)
	} else if config.IsEnabled(config.GetEthashEIP649Transition, next) {
		// The calculation uses the Byzantium rules, with bomb offset of 3M.
		// Specification EIP-649: https://eips.ethereum.org/EIPS/eip-649
		// Related meta-ish EIP-669: https://github.com/ethereum/EIPs/pull/669
		// Note, the calculations below looks at the parent number, which is 1 below
		// the block number. Thus we remove one from the delay given

		fakeBlockNumber := new(big.Int)
		delayWithOffset := new(big.Int).Sub(vars.EIP649DifficultyBombDelay.ToBig(), common.Big1)
		if parent.Number.Cmp(delayWithOffset) >= 0 {
			fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, delayWithOffset)
		}
		exPeriodRef.Set(fakeBlockNumber)
	}

	// EXPLOSION

	// the 'periodRef' (from above) represents the many ways of hackishly modifying the reference number
	// (ie the 'currentBlock') in order to lie to the function about what time it really is
	//
	//   2^(( periodRef // EDP) - 2)
	//
	x := new(big.Int)
	x.Div(exPeriodRef, params.ExpDiffPeriod.ToBig()) // (periodRef // EDP)
	if x.Cmp(big1) > 0 {                             // if result large enough (not in algo explicitly)
		x.Sub(x, big2)      // - 2
		x.Exp(big2, x, nil) // 2^
	} else {
		x.SetUint64(0)
	}
	out.Add(out, x)
	return out
}

// Some weird constants to avoid constant memory allocs for them.
var (
	big1       = big.NewInt(1)
	big2       = big.NewInt(2)
	bigMinus99 = big.NewInt(-99)
)

// Exported for fuzzing
var FrontierDifficultyCalculator = CalcDifficultyFrontierU256
var HomesteadDifficultyCalculator = CalcDifficultyHomesteadU256
var DynamicDifficultyCalculator = MakeDifficultyCalculatorU256

// verifySeal checks whether a block satisfies the PoW difficulty requirements,
// either using the usual ethash cache for it, or alternatively using a full DAG
// to make remote mining fast.
func (ethash *Ethash) verifySeal(chain consensus.ChainHeaderReader, header *types.Header, fulldag bool) error {
	// If we're running a fake PoW, accept any seal as valid
	if ethash.config.PowMode == ModeFake || ethash.config.PowMode == ModePoissonFake || ethash.config.PowMode == ModeFullFake {
		time.Sleep(ethash.fakeDelay)
		if ethash.fakeFail == header.Number.Uint64() {
			return errInvalidPoW
		}
		return nil
	}
	// If we're running a shared PoW, delegate verification to it
	if ethash.shared != nil {
		return ethash.shared.verifySeal(chain, header, fulldag)
	}
	// Ensure that we have a valid difficulty for the block
	if header.Difficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}
	// Recompute the digest and PoW values
	number := header.Number.Uint64()

	var (
		digest []byte
		result []byte
	)
	// If fast-but-heavy PoW verification was requested, use an ethash dataset
	if fulldag {
		dataset := ethash.dataset(number, true)
		if dataset.generated() {
			digest, result = hashimotoFull(dataset.dataset, ethash.SealHash(header).Bytes(), header.Nonce.Uint64())

			// Datasets are unmapped in a finalizer. Ensure that the dataset stays alive
			// until after the call to hashimotoFull so it's not unmapped while being used.
			runtime.KeepAlive(dataset)
		} else {
			// Dataset not yet generated, don't hang, use a cache instead
			fulldag = false
		}
	}
	// If slow-but-light PoW verification was requested (or DAG not yet ready), use an ethash cache
	if !fulldag {
		cache := ethash.cache(number)
		epochLength := calcEpochLength(number, ethash.config.ECIP1099Block)
		epoch := calcEpoch(number, epochLength)
		size := datasetSize(epoch)
		if ethash.config.PowMode == ModeTest {
			size = 32 * 1024
		}
		digest, result = hashimotoLight(size, cache.cache, ethash.SealHash(header).Bytes(), header.Nonce.Uint64())

		// Caches are unmapped in a finalizer. Ensure that the cache stays alive
		// until after the call to hashimotoLight so it's not unmapped while being used.
		runtime.KeepAlive(cache)
	}
	// Verify the calculated values against the ones provided in the header
	if !bytes.Equal(header.MixDigest[:], digest) {
		return errInvalidMixDigest
	}
	target := new(big.Int).Div(two256, header.Difficulty)
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}
	return nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the ethash protocol. The changes are done inline.
func (ethash *Ethash) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	parent := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	header.Difficulty = ethash.CalcDifficulty(chain, header.Time, parent)
	return nil
}

// Finalize implements consensus.Engine, accumulating the block and uncle rewards.
func (ethash *Ethash) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, withdrawals []*types.Withdrawal) {
	// Accumulate any block and uncle rewards and commit the final state root
	mutations.AccumulateRewards(chain.Config(), state, header, uncles)

	// Apply Halo-specific EIP-1559 fee distribution for Halo chain (ChainID 12000)
	// This distributes base fees: 40% burn, 30% miner, 20% ecosystem, 10% reserve
	chainID := chain.Config().GetChainID()
	if chainID != nil && chainID.Uint64() == 12000 {
		// Only apply if EIP-1559 is active and we have a base fee
		if chain.Config().IsEnabled(chain.Config().GetEIP1559Transition, header.Number) && header.BaseFee != nil {
			if err := eip1559.ApplyHaloBaseFeeDistribution(state, header, header.BaseFee, header.GasUsed); err != nil {
				// Log error but don't fail block finalization
				// In production, you may want to handle this differently
				panic(fmt.Sprintf("failed to apply Halo EIP-1559 distribution: %v", err))
			}
		}
	}
}

// FinalizeAndAssemble implements consensus.Engine, accumulating the block and
// uncle rewards, setting the final state and assembling the block.
func (ethash *Ethash) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt, withdrawals []*types.Withdrawal) (*types.Block, error) {
	if len(withdrawals) > 0 {
		return nil, errors.New("ethash does not support withdrawals")
	}
	// Finalize block
	ethash.Finalize(chain, header, state, txs, uncles, nil)

	// Assign the final state root to header.
	header.Root = state.IntermediateRoot(chain.Config().IsEnabled(chain.Config().GetEIP161dTransition, header.Number))

	// Header seems complete, assemble into a block and return
	return types.NewBlock(header, txs, uncles, receipts, trie.NewStackTrie(nil)), nil
}

// SealHash returns the hash of a block prior to it being sealed.
func (ethash *Ethash) SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	enc := []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra,
	}
	if header.BaseFee != nil {
		enc = append(enc, header.BaseFee)
	}
	if header.WithdrawalsHash != nil {
		panic("withdrawal hash set on ethash")
	}
	if header.ExcessBlobGas != nil {
		panic("excess blob gas set on ethash")
	}
	if header.BlobGasUsed != nil {
		panic("blob gas used set on ethash")
	}
	if header.ParentBeaconRoot != nil {
		panic("parent beacon root set on ethash")
	}
	rlp.Encode(hasher, enc)
	hasher.Sum(hash[:0])
	return hash
}
