package params_test

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/params"
)

func TestDefaultHaloGenesisBlock(t *testing.T) {
	genesisBlock := params.DefaultHaloGenesisBlock()

	// Verify basic parameters
	if genesisBlock.GasLimit != 150000000 {
		t.Fatalf("expected gas limit 150000000, got %d", genesisBlock.GasLimit)
	}

	if genesisBlock.Config.GetChainID().Uint64() != 12000 {
		t.Fatalf("expected chain ID 12000, got %d", genesisBlock.Config.GetChainID().Uint64())
	}

	// Print genesis JSON for reference
	genesisBlockJSON, err := json.MarshalIndent(genesisBlock, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Halo Genesis Block JSON:")
	t.Log(string(genesisBlockJSON))
}

func TestHaloGenesisExtraData(t *testing.T) {
	genesisBlock := params.DefaultHaloGenesisBlock()

	// Verify extra data is set (should be "Halo Network")
	if len(genesisBlock.ExtraData) == 0 {
		t.Fatal("expected extra data to be set")
	}

	t.Logf("Genesis ExtraData: %x", genesisBlock.ExtraData)
}
