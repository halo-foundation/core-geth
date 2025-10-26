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

package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
)

var (
	// HaloChainConfig is the chain parameters to run a node on the Halo network.
	// Chain ID: 12000
	// Consensus: Proof of Work (Ethash with custom parameters)
	// Block time: 1 second
	// EIP-1559: Enabled with custom fee distribution
	HaloChainConfig = &coregeth.CoreGethChainConfig{
		NetworkID: 12000,
		ChainID:   big.NewInt(12000),

		// Custom Ethash configuration for Halo
		// Note: DAG parameters and difficulty adjustment are configured via vars/halo_vars.go
		// and used in the consensus engine, not in the config struct
		Ethash: &ctypes.EthashConfig{},

		// Homestead (EIP-2, EIP-7)
		EIP2FBlock: big.NewInt(0),
		EIP7FBlock: big.NewInt(0),

		// Tangerine Whistle (EIP-150)
		EIP150Block: big.NewInt(0),

		// Spurious Dragon (EIP-155, EIP-160, EIP-161, EIP-170)
		EIP155Block:  big.NewInt(0),
		EIP160FBlock: big.NewInt(0),
		EIP161FBlock: big.NewInt(0),
		EIP170FBlock: big.NewInt(0),

		// Byzantium (EIP-100, EIP-140, EIP-198, EIP-211, EIP-212, EIP-213, EIP-214, EIP-658)
		EIP100FBlock: big.NewInt(0),
		EIP140FBlock: big.NewInt(0),
		EIP198FBlock: big.NewInt(0),
		EIP211FBlock: big.NewInt(0),
		EIP212FBlock: big.NewInt(0),
		EIP213FBlock: big.NewInt(0),
		EIP214FBlock: big.NewInt(0),
		EIP658FBlock: big.NewInt(0),

		// Constantinople (EIP-145, EIP-1014, EIP-1052)
		EIP145FBlock:  big.NewInt(0),
		EIP1014FBlock: big.NewInt(0),
		EIP1052FBlock: big.NewInt(0),

		// Istanbul (EIP-152, EIP-1108, EIP-1344, EIP-1884, EIP-2028, EIP-2200)
		EIP152FBlock:  big.NewInt(0),
		EIP1108FBlock: big.NewInt(0),
		EIP1344FBlock: big.NewInt(0),
		EIP1884FBlock: big.NewInt(0),
		EIP2028FBlock: big.NewInt(0),
		EIP2200FBlock: big.NewInt(0),

		// Berlin (EIP-2565, EIP-2718, EIP-2929, EIP-2930)
		EIP2565FBlock: big.NewInt(0),
		EIP2718FBlock: big.NewInt(0),
		EIP2929FBlock: big.NewInt(0),
		EIP2930FBlock: big.NewInt(0),

		// London (EIP-1559, EIP-3198, EIP-3529, EIP-3541)
		// Custom EIP-1559 implementation with 4-way fee split
		EIP1559FBlock: big.NewInt(0),
		EIP3198FBlock: big.NewInt(0),
		EIP3529FBlock: big.NewInt(0),
		EIP3541FBlock: big.NewInt(0),

		// Shanghai (EIP-3651, EIP-3855, EIP-3860)
		EIP3651FBlock: big.NewInt(0),
		EIP3855FBlock: big.NewInt(0),
		EIP3860FBlock: big.NewInt(0),

		// Halo-specific configurations
		// Block rewards: 5 HALO until block 100,000
		// Then decreases by 1 HALO every 300,000 blocks
		// At block 1,000,000: 2 HALO per block
		DisposalBlock: big.NewInt(0), // Defuse difficulty bomb from genesis

		// Uncle rewards depth configuration
		// MaxUncles = 1 (enforced in consensus)
		// MaxUnclesDepth = 2
		// UncleRewardDepth1 = 875/1000 of block reward
		// UncleRewardDepth2 = 750/1000 of block reward
		// NephewReward = 31/1000 of block reward per uncle

		RequireBlockHashes: map[uint64]common.Hash{},
	}
)
