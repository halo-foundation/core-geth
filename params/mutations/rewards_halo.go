// Copyright 2025 The core-geth Authors
// This file is part of the core-geth library.
//
// The core-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The core-geth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the core-geth library. If not, see <http://www.gnu.org/licenses/>.

package mutations

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

// Halo reward constants - 100M Max Supply with Phased Distribution
var (
	// Phase milestones (in blocks)
	haloPhase1End = uint64(25000)        // 1M tokens @ 40 HALO/block
	haloPhase2End = uint64(358333)       // 2M tokens @ 3 HALO/block (25k + 333,333)
	haloPhase3End = uint64(691666)       // 2.5M tokens @ 1.5 HALO/block
	haloPhase4End = uint64(5691666)      // 7.5M tokens @ 1 HALO/block

	// Year boundaries (1 block/sec = 31,536,000 blocks/year)
	haloYear1End = uint64(31536000)      // End of Year 1
	haloYear2End = uint64(63072000)      // End of Year 2
	haloYear3End = uint64(94608000)      // End of Year 3
	haloYear4End = uint64(126144000)     // End of Year 4
	haloYear5End = uint64(157680000)     // End of Year 5
	haloYear6End = uint64(189216000)     // End of Year 6

	// Block rewards (in wei) - 1 HALO = 10^18 wei
	haloPhase1Reward = new(uint256.Int).Mul(uint256.NewInt(40), uint256.NewInt(1e18))  // 40 HALO
	haloPhase2Reward = new(uint256.Int).Mul(uint256.NewInt(3), uint256.NewInt(1e18))   // 3 HALO
	haloPhase3Reward = new(uint256.Int).Mul(uint256.NewInt(15), uint256.NewInt(1e17))  // 1.5 HALO
	haloPhase4Reward = uint256.NewInt(1e18)   // 1 HALO
	haloPhase5Reward = uint256.NewInt(5e17)   // 0.5 HALO

	// Halving rewards (multiply by 75% each year)
	haloYear2Reward = uint256.NewInt(375e15)      // 0.375 HALO (0.5 * 0.75)
	haloYear3Reward = uint256.NewInt(28125e13)    // 0.28125 HALO (0.375 * 0.75)
	haloYear4Reward = uint256.NewInt(2109375e11)  // 0.2109375 HALO
	haloYear5Reward = uint256.NewInt(158203125e9) // 0.158203125 HALO
	haloMinimumBlockReward = uint256.NewInt(125e15) // 0.125 HALO minimum

	// Maximum supply: 100 million HALO
	haloMaxSupply = new(uint256.Int).Mul(uint256.NewInt(100000000), uint256.NewInt(1e18))

	// Uncle reward ratios (per 1000)
	haloUncleRewardDepth1 = uint256.NewInt(875) // 87.5% of block reward for depth 1
	haloUncleRewardDepth2 = uint256.NewInt(750) // 75% of block reward for depth 2
	haloNephewReward      = uint256.NewInt(31)  // 3.1% of block reward per uncle
	haloDenominator       = uint256.NewInt(1000)

	haloOne = uint256.NewInt(1)
)

// GetHaloBlockReward calculates the block reward for a given block number
// Implements phased reduction schedule with 100M max supply:
//
// Phase 1 (blocks 0-25,000): 40 HALO → 1M tokens in 6.94 hours
// Phase 2 (blocks 25,001-358,333): 3 HALO → 1M tokens in 3.86 days
// Phase 3 (blocks 358,334-691,666): 1.5 HALO → 0.5M tokens in 3.86 days
// Phase 4 (blocks 691,667-5,691,666): 1 HALO → 5M tokens in 57.87 days
// Total Phases 1-4: 7.5M tokens in ~65.9 days
//
// Phase 5 (Year 1 remainder): 0.5 HALO → ~12.92M tokens
// Year 1 total: ~20.42M tokens
//
// Year 2: 0.375 HALO (75% of previous) → 11.826M tokens (total: 32.248M)
// Year 3: 0.28125 HALO → 8.870M tokens (total: 41.118M)
// Year 4: 0.2109375 HALO → 6.652M tokens (total: 47.770M)
// Year 5: 0.158203125 HALO → 4.989M tokens (total: 52.759M)
// Year 6+: 0.125 HALO (minimum floor) → ~3.942M tokens/year
//
// Max supply of 100M reached in approximately 18 years
func GetHaloBlockReward(blockNum *big.Int) *uint256.Int {
	blockNumber := blockNum.Uint64()

	// Phase 1: 0-25,000 blocks → 40 HALO/block
	if blockNumber < haloPhase1End {
		return new(uint256.Int).Set(haloPhase1Reward)
	}

	// Phase 2: 25,001-358,333 blocks → 3 HALO/block
	if blockNumber < haloPhase2End {
		return new(uint256.Int).Set(haloPhase2Reward)
	}

	// Phase 3: 358,334-691,666 blocks → 1.5 HALO/block
	if blockNumber < haloPhase3End {
		return new(uint256.Int).Set(haloPhase3Reward)
	}

	// Phase 4: 691,667-5,691,666 blocks → 1 HALO/block
	if blockNumber < haloPhase4End {
		return new(uint256.Int).Set(haloPhase4Reward)
	}

	// Phase 5: Remainder of Year 1 → 0.5 HALO/block
	if blockNumber < haloYear1End {
		return new(uint256.Int).Set(haloPhase5Reward)
	}

	// Year 2: 0.375 HALO/block (75% halving)
	if blockNumber < haloYear2End {
		return new(uint256.Int).Set(haloYear2Reward)
	}

	// Year 3: 0.28125 HALO/block
	if blockNumber < haloYear3End {
		return new(uint256.Int).Set(haloYear3Reward)
	}

	// Year 4: 0.2109375 HALO/block
	if blockNumber < haloYear4End {
		return new(uint256.Int).Set(haloYear4Reward)
	}

	// Year 5: 0.158203125 HALO/block
	if blockNumber < haloYear5End {
		return new(uint256.Int).Set(haloYear5Reward)
	}

	// Year 6+: 0.125 HALO/block (minimum floor)
	return new(uint256.Int).Set(haloMinimumBlockReward)
}

// GetHaloUncleReward calculates the uncle reward based on depth
// Depth 1 (uncle is 1 block behind): 87.5% of block reward
// Depth 2 (uncle is 2 blocks behind): 75% of block reward
func GetHaloUncleReward(header, uncle *types.Header, blockReward *uint256.Int) *uint256.Int {
	// Calculate depth
	depth := new(uint256.Int).Sub(uint256.MustFromBig(header.Number), uint256.MustFromBig(uncle.Number))

	var rewardRatio *uint256.Int
	if depth.Cmp(haloOne) == 0 {
		// Depth 1: 87.5%
		rewardRatio = haloUncleRewardDepth1
	} else if depth.Cmp(uint256.NewInt(2)) == 0 {
		// Depth 2: 75%
		rewardRatio = haloUncleRewardDepth2
	} else {
		// No reward for depth > 2
		return uint256.NewInt(0)
	}

	// Calculate: blockReward * ratio / 1000
	reward := new(uint256.Int).Mul(blockReward, rewardRatio)
	reward.Div(reward, haloDenominator)

	return reward
}

// GetHaloNephewReward calculates the reward for the block miner for including uncles
// Miner gets 3.1% of block reward per uncle (max 1 uncle)
func GetHaloNephewReward(uncles []*types.Header, blockReward *uint256.Int) *uint256.Int {
	if len(uncles) == 0 {
		return uint256.NewInt(0)
	}

	// 3.1% per uncle
	rewardPerUncle := new(uint256.Int).Mul(blockReward, haloNephewReward)
	rewardPerUncle.Div(rewardPerUncle, haloDenominator)

	// Multiply by number of uncles (max 1 for Halo)
	totalReward := new(uint256.Int).Mul(rewardPerUncle, uint256.NewInt(uint64(len(uncles))))

	return totalReward
}

// haloBlockReward calculates the Halo network block rewards
// Returns (minerReward, uncleRewards[])
func haloBlockReward(header *types.Header, uncles []*types.Header) (*uint256.Int, []*uint256.Int) {
	blockReward := GetHaloBlockReward(header.Number)

	// Calculate uncle rewards
	uncleRewards := make([]*uint256.Int, len(uncles))
	for i, uncle := range uncles {
		uncleRewards[i] = GetHaloUncleReward(header, uncle, blockReward)
	}

	// Calculate miner reward: base reward + nephew rewards
	minerReward := new(uint256.Int).Set(blockReward)
	nephewReward := GetHaloNephewReward(uncles, blockReward)
	minerReward.Add(minerReward, nephewReward)

	return minerReward, uncleRewards
}
