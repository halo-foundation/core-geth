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

package vars

import (
	"math/big"
)

// Halo network consensus parameters
var (
	// Block time and gas limits
	HaloTargetBlockTime = big.NewInt(1)          // 1 second block time
	HaloGenesisGasLimit = uint64(150000000)      // 150M genesis gas limit
	HaloMinGasLimit     = uint64(50000000)       // 50M minimum gas limit
	HaloMaxGasLimit     = uint64(300000000)      // 300M maximum gas limit
	HaloGasLimitBoundDivisor = uint64(2048)      // Gas limit adjustment divisor

	// Difficulty parameters
	HaloDifficultyBoundDivisor = big.NewInt(2048) // 2048 difficulty bound divisor
	HaloDurationLimit          = big.NewInt(1)    // 1 second duration limit

	// Uncle parameters
	HaloMaxUncles      = 1                        // Maximum 1 uncle per block
	HaloMaxUnclesDepth = uint64(2)                // Uncles can be max 2 blocks deep

	// EIP-1559 parameters
	HaloInitialBaseFee              = uint64(1000000000) // 1 Gwei initial base fee
	HaloBaseFeeChangeDenominator    = uint64(8)          // Base fee change denominator
	HaloElasticityMultiplier        = uint64(2)          // Elasticity multiplier

	// Cache and database settings
	HaloStateCacheSize      = 1000000              // 1M state cache entries
	HaloCodeCacheSize       = 100000               // 100K code cache entries
	HaloTrieCleanCacheSize  = 512                  // 512 MB trie clean cache
	HaloTrieDirtyCacheSize  = 256                  // 256 MB trie dirty cache
	HaloDatabaseCache       = 2048                 // 2 GB database cache

	// Transaction pool settings
	HaloGlobalSlots     = uint64(8192)             // 8192 global transaction slots
	HaloGlobalQueue     = uint64(4096)             // 4096 global queue slots
	HaloAccountSlots    = uint64(128)              // 128 slots per account
	HaloAccountQueue    = uint64(64)               // 64 queue slots per account

	// Network settings
	HaloMaxMessageSize  = 10 * 1024 * 1024         // 10 MB max message size
	HaloMaxPeers        = 50                       // 50 max peers
)

// Halo DAG parameters for Ethash
var (
	// Custom DAG size: 512 MB initial
	HaloInitialDAGSize = uint64(512 * 1024 * 1024)  // 512 MB

	// Custom DAG growth: 1 MB per epoch
	HaloDAGGrowth = uint64(1 * 1024 * 1024)         // 1 MB

	// Epoch length (standard Ethash: 30000 blocks)
	HaloEpochLength = uint64(30000)                 // 30000 blocks per epoch
)
