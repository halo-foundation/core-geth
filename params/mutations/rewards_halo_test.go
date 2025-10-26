// Copyright 2025 The core-geth Authors
// This file is part of the core-geth library.

package mutations

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

func TestGetHaloBlockReward(t *testing.T) {
	tests := []struct {
		blockNum uint64
		expected string // in HALO (will be multiplied by 1e18)
	}{
		// Phase 1: 0-25,000 blocks → 40 HALO/block
		{0, "40"},
		{10000, "40"},
		{24999, "40"},

		// Phase 2: 25,001-358,333 blocks → 3 HALO/block
		{25000, "3"},
		{100000, "3"},
		{358332, "3"},

		// Phase 3: 358,334-691,666 blocks → 1.5 HALO/block
		{358333, "1.5"},
		{500000, "1.5"},
		{691665, "1.5"},

		// Phase 4: 691,667-5,691,666 blocks → 1 HALO/block
		{691666, "1"},
		{1000000, "1"},
		{5691665, "1"},

		// Phase 5: Remainder of Year 1 → 0.5 HALO/block
		{5691666, "0.5"},
		{10000000, "0.5"},
		{31535999, "0.5"},

		// Year 2: 0.375 HALO/block
		{31536000, "0.375"},
		{40000000, "0.375"},
		{63071999, "0.375"},

		// Year 3: 0.28125 HALO/block
		{63072000, "0.28125"},
		{80000000, "0.28125"},
		{94607999, "0.28125"},

		// Year 4: 0.2109375 HALO/block
		{94608000, "0.2109375"},
		{100000000, "0.2109375"},
		{126143999, "0.2109375"},

		// Year 5: 0.158203125 HALO/block
		{126144000, "0.158203125"},
		{140000000, "0.158203125"},
		{157679999, "0.158203125"},

		// Year 6+: 0.125 HALO/block (minimum floor)
		{157680000, "0.125"},
		{200000000, "0.125"},
		{500000000, "0.125"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.blockNum)), func(t *testing.T) {
			reward := GetHaloBlockReward(big.NewInt(int64(tt.blockNum)))

			// Convert expected to wei
			expectedHalo := new(big.Float)
			expectedHalo.SetString(tt.expected)
			weiPerHalo := new(big.Float).SetInt(big.NewInt(1e18))
			expectedWei := new(big.Float).Mul(expectedHalo, weiPerHalo)

			expectedInt := new(big.Int)
			expectedWei.Int(expectedInt)
			expectedU256 := uint256.MustFromBig(expectedInt)

			if reward.Cmp(expectedU256) != 0 {
				t.Errorf("Block %d: expected %s HALO (%s wei), got %s wei",
					tt.blockNum, tt.expected, expectedU256.ToBig().String(), reward.ToBig().String())
			}
		})
	}
}

func TestGetHaloUncleReward(t *testing.T) {
	// Test uncle rewards at different depths with Phase 1 reward
	baseReward := new(uint256.Int).Mul(uint256.NewInt(40), uint256.NewInt(1e18)) // 40 HALO base reward

	// Create header at block 1000 (Phase 1)
	header := &types.Header{
		Number: big.NewInt(1000),
	}

	tests := []struct {
		name         string
		uncleBlock   uint64
		expectedPct  string // Percentage of base reward
	}{
		{"Depth 1", 999, "87.5"},  // 87.5% of 40 HALO
		{"Depth 2", 998, "75.0"},  // 75% of 40 HALO
		{"Depth 3", 997, "0"},     // No reward for depth > 2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uncle := &types.Header{
				Number: big.NewInt(int64(tt.uncleBlock)),
			}

			reward := GetHaloUncleReward(header, uncle, baseReward)

			// Calculate expected reward
			var expected *uint256.Int
			if tt.expectedPct == "87.5" {
				// 87.5% = 875/1000
				expected = new(uint256.Int).Mul(baseReward, uint256.NewInt(875))
				expected.Div(expected, uint256.NewInt(1000))
			} else if tt.expectedPct == "75.0" {
				// 75% = 750/1000
				expected = new(uint256.Int).Mul(baseReward, uint256.NewInt(750))
				expected.Div(expected, uint256.NewInt(1000))
			} else {
				expected = uint256.NewInt(0)
			}

			if reward.Cmp(expected) != 0 {
				t.Errorf("Expected %s%% of base reward (%s), got %s",
					tt.expectedPct, expected.ToBig().String(), reward.ToBig().String())
			}
		})
	}
}

func TestGetHaloNephewReward(t *testing.T) {
	baseReward := new(uint256.Int).Mul(uint256.NewInt(40), uint256.NewInt(1e18)) // 40 HALO base reward

	tests := []struct {
		name         string
		uncleCount   int
		expectedPct  string
	}{
		{"No uncles", 0, "0"},
		{"1 uncle", 1, "3.1"},  // 3.1% of 40 HALO
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create dummy uncles
			uncles := make([]*types.Header, tt.uncleCount)
			for i := 0; i < tt.uncleCount; i++ {
				uncles[i] = &types.Header{}
			}

			reward := GetHaloNephewReward(uncles, baseReward)

			// Calculate expected reward: 3.1% per uncle
			expected := new(uint256.Int).Mul(baseReward, uint256.NewInt(31))
			expected.Div(expected, uint256.NewInt(1000))
			expected.Mul(expected, uint256.NewInt(uint64(tt.uncleCount)))

			if reward.Cmp(expected) != 0 {
				t.Errorf("Expected %s%% of base reward (%s), got %s",
					tt.expectedPct, expected.ToBig().String(), reward.ToBig().String())
			}
		})
	}
}

func TestHaloBlockReward(t *testing.T) {
	// Test complete block reward calculation with uncles in Phase 1

	header := &types.Header{
		Number: big.NewInt(10000), // Block 10000: 40 HALO base reward (Phase 1)
	}

	uncle := &types.Header{
		Number: big.NewInt(9999), // Depth 1 uncle
	}

	uncles := []*types.Header{uncle}

	minerReward, uncleRewards := haloBlockReward(header, uncles)

	// Expected uncle reward: 87.5% of 40 HALO
	baseReward := new(uint256.Int).Mul(uint256.NewInt(40), uint256.NewInt(1e18))
	expectedUncleReward := new(uint256.Int).Mul(baseReward, uint256.NewInt(875))
	expectedUncleReward.Div(expectedUncleReward, uint256.NewInt(1000))

	if uncleRewards[0].Cmp(expectedUncleReward) != 0 {
		t.Errorf("Uncle reward mismatch: expected %s, got %s",
			expectedUncleReward.ToBig().String(), uncleRewards[0].ToBig().String())
	}

	// Expected miner reward: 40 HALO + 3.1% nephew reward
	nephewReward := new(uint256.Int).Mul(baseReward, uint256.NewInt(31))
	nephewReward.Div(nephewReward, uint256.NewInt(1000))

	expectedMinerReward := new(uint256.Int).Add(baseReward, nephewReward)

	if minerReward.Cmp(expectedMinerReward) != 0 {
		t.Errorf("Miner reward mismatch: expected %s, got %s",
			expectedMinerReward.ToBig().String(), minerReward.ToBig().String())
	}

	t.Logf("Block 10000 rewards (Phase 1):")
	t.Logf("  Miner: %s wei (%.6f HALO)", minerReward.ToBig().String(), float64(minerReward.ToBig().Uint64())/1e18)
	t.Logf("  Uncle: %s wei (%.6f HALO)", uncleRewards[0].ToBig().String(), float64(uncleRewards[0].ToBig().Uint64())/1e18)
}

func TestHaloTokenomics(t *testing.T) {
	// Test that the tokenomics schedule reaches approximately 100M tokens

	type phase struct {
		name       string
		startBlock uint64
		endBlock   uint64
		reward     string
	}

	phases := []phase{
		{"Phase 1", 0, 25000, "40"},
		{"Phase 2", 25000, 358333, "3"},
		{"Phase 3", 358333, 691666, "1.5"},
		{"Phase 4", 691666, 5691666, "1"},
		{"Phase 5", 5691666, 31536000, "0.5"},
		{"Year 2", 31536000, 63072000, "0.375"},
		{"Year 3", 63072000, 94608000, "0.28125"},
		{"Year 4", 94608000, 126144000, "0.2109375"},
		{"Year 5", 126144000, 157680000, "0.158203125"},
	}

	totalTokens := new(big.Float)

	for _, p := range phases {
		blocks := p.endBlock - p.startBlock
		rewardPerBlock := new(big.Float)
		rewardPerBlock.SetString(p.reward)
		phaseTokens := new(big.Float).Mul(rewardPerBlock, big.NewFloat(float64(blocks)))
		totalTokens.Add(totalTokens, phaseTokens)

		t.Logf("%s: %d blocks × %s HALO = %.2f HALO",
			p.name, blocks, p.reward, phaseTokens)
	}

	// Add minimum floor tokens for remaining years to reach ~100M
	// Year 6 onwards: 0.125 HALO/block
	// To reach 100M, we need: 100M - totalSoFar
	remainingTokens := new(big.Float).Sub(big.NewFloat(100000000), totalTokens)
	blocksNeeded := new(big.Float).Quo(remainingTokens, big.NewFloat(0.125))

	t.Logf("\nTotal after Year 5: %.2f HALO", totalTokens)
	t.Logf("Remaining to 100M: %.2f HALO", remainingTokens)
	t.Logf("Blocks needed at 0.125 HALO/block: %.0f", blocksNeeded)

	yearsNeeded := new(big.Float).Quo(blocksNeeded, big.NewFloat(31536000))
	t.Logf("Years to reach 100M supply: %.1f years after Year 5", yearsNeeded)

	// Verify we're targeting 100M max supply
	maxSupply := haloMaxSupply.ToBig()
	expectedMaxSupply := new(big.Int).Mul(big.NewInt(100000000), big.NewInt(1e18))

	if maxSupply.Cmp(expectedMaxSupply) != 0 {
		t.Errorf("Max supply mismatch: expected %s, got %s",
			expectedMaxSupply.String(), maxSupply.String())
	}
}
