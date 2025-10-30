package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params/types/genesisT"
)

// HaloGenesisHash - This will be updated after first genesis initialization
var HaloGenesisHash = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")

// Halo fund addresses
var (
	// EcosystemFundAddress receives 20% of transaction fees
	// TODO: Replace with actual address before deployment
	HaloEcosystemFundAddress = common.HexToAddress("0xa7548DF196e2C1476BDc41602E288c0A8F478c4f")

	// ReserveFundAddress receives 10% of transaction fees
	// TODO: Replace with actual address before deployment
	HaloReserveFundAddress = common.HexToAddress("0xb95ae9b737e104C666d369CFb16d6De88208Bd80")
)

// DefaultHaloGenesisBlock returns the Halo network genesis block.
// UPDATED for 4-second block time with enhanced security
func DefaultHaloGenesisBlock() *genesisT.Genesis {
	return &genesisT.Genesis{
		Config:     HaloChainConfig,
		Nonce:      0,
		ExtraData:  hexutil.MustDecode("0x48616c6f204e6574776f726b20763120343273"), // "Halo Network v1 4s" in hex
		GasLimit:   150000000, // 150M gas limit (GenesisGasLimit)
		Difficulty: hexutil.MustDecodeBig("0x1F4"), // 500 - very low for easy testnet bootstrap
		Timestamp:  1700000000, // TODO: Set to actual launch timestamp
		Alloc: genesisT.GenesisAlloc{
			// Ecosystem Fund - receives 20% of fees
			HaloEcosystemFundAddress: {
				Balance: big.NewInt(0), // No pre-mine, funded through fees
			},
			// Reserve Fund - receives 10% of fees
			HaloReserveFundAddress: {
				Balance: big.NewInt(0), // No pre-mine, funded through fees
			},
			// TODO: Add any pre-mine allocations here
			// Example:
			// common.HexToAddress("0x..."): {
			//     Balance: new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18)), // 1M HALO
			// },
		},
	}
}
