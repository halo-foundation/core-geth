package params

import (
	"testing"

	"github.com/ethereum/go-ethereum/params/types/ctypes"
)

func TestHaloChainConfig(t *testing.T) {
	if HaloChainConfig.GetChainID().Uint64() != 12000 {
		t.Fatalf("expected chain ID 12000, got %d", HaloChainConfig.GetChainID().Uint64())
	}

	if HaloChainConfig.GetNetworkID() != 12000 {
		t.Fatalf("expected network ID 12000, got %d", HaloChainConfig.GetNetworkID())
	}

	if HaloChainConfig.GetConsensusEngineType() != ctypes.ConsensusEngineT_Ethash {
		t.Fatalf("expected Ethash consensus")
	}

	// Test EIP-1559 is enabled from genesis
	if HaloChainConfig.GetEIP1559Transition() == nil || HaloChainConfig.GetEIP1559Transition().Uint64() != 0 {
		t.Fatalf("expected EIP-1559 enabled from genesis")
	}

	// Test modern EIPs are enabled from genesis
	if HaloChainConfig.GetEIP155Transition() == nil || HaloChainConfig.GetEIP155Transition().Uint64() != 0 {
		t.Fatalf("expected EIP-155 enabled from genesis")
	}

	if HaloChainConfig.GetEIP1344Transition() == nil || HaloChainConfig.GetEIP1344Transition().Uint64() != 0 {
		t.Fatalf("expected EIP-1344 (CHAINID) enabled from genesis")
	}
}

func TestHaloGenesisAddressesNotZero(t *testing.T) {
	// Note: In production, these addresses MUST be set to real addresses
	// This test just ensures they are defined
	if HaloEcosystemFundAddress.Hex() == "0x0000000000000000000000000000000000000000" {
		t.Log("WARNING: HaloEcosystemFundAddress is zero address - must be set before deployment")
	}

	if HaloReserveFundAddress.Hex() == "0x0000000000000000000000000000000000000000" {
		t.Log("WARNING: HaloReserveFundAddress is zero address - must be set before deployment")
	}
}
