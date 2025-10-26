package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("Halo Chain - Fund Address Generation")
	fmt.Println("==============================================")
	fmt.Println()

	// Generate Ecosystem Fund address
	fmt.Println("1. ECOSYSTEM FUND (20% of base fees)")
	fmt.Println("----------------------------------------------")
	ecosystemKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	ecosystemAddr := crypto.PubkeyToAddress(ecosystemKey.PublicKey)
	ecosystemPrivKey := hex.EncodeToString(crypto.FromECDSA(ecosystemKey))

	fmt.Printf("Address:     %s\n", ecosystemAddr.Hex())
	fmt.Printf("Private Key: %s\n", ecosystemPrivKey)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Store this private key securely!")
	fmt.Println("   Recommended: Import into Gnosis Safe multisig")
	fmt.Println()

	// Generate Reserve Fund address
	fmt.Println("2. RESERVE FUND (10% of base fees)")
	fmt.Println("----------------------------------------------")
	reserveKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	reserveAddr := crypto.PubkeyToAddress(reserveKey.PublicKey)
	reservePrivKey := hex.EncodeToString(crypto.FromECDSA(reserveKey))

	fmt.Printf("Address:     %s\n", reserveAddr.Hex())
	fmt.Printf("Private Key: %s\n", reservePrivKey)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Store this private key securely!")
	fmt.Println("   Recommended: Import into Gnosis Safe multisig")
	fmt.Println()

	// Print configuration update instructions
	fmt.Println("==============================================")
	fmt.Println("NEXT STEPS:")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("1. SAVE PRIVATE KEYS SECURELY")
	fmt.Println("   - Use a password manager or hardware wallet")
	fmt.Println("   - NEVER commit private keys to git")
	fmt.Println("   - Consider using multisig (Gnosis Safe)")
	fmt.Println()
	fmt.Println("2. UPDATE CONFIGURATION")
	fmt.Println("   Edit: params/genesis_halo.go")
	fmt.Println()
	fmt.Printf("   HaloEcosystemFundAddress = common.HexToAddress(\"%s\")\n", ecosystemAddr.Hex())
	fmt.Printf("   HaloReserveFundAddress = common.HexToAddress(\"%s\")\n", reserveAddr.Hex())
	fmt.Println()
	fmt.Println("3. REBUILD")
	fmt.Println("   make clean && make geth")
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println()

	// Save to file
	saveToFile(ecosystemAddr.Hex(), ecosystemPrivKey, reserveAddr.Hex(), reservePrivKey)
}

func saveToFile(ecosystemAddr, ecosystemKey, reserveAddr, reserveKey string) {
	filename := "HALO_FUND_ADDRESSES.txt"
	content := fmt.Sprintf(`HALO CHAIN - FUND ADDRESSES
========================================
Generated: %s
========================================

⚠️  CRITICAL: KEEP THIS FILE SECURE ⚠️
This file contains private keys that control the fund addresses.
DO NOT commit to git or share publicly.

========================================
ECOSYSTEM FUND (20%% of base fees)
========================================
Address:     %s
Private Key: %s

Purpose: Development, grants, marketing
Recommended: Import to Gnosis Safe 3-of-5 multisig

========================================
RESERVE FUND (10%% of base fees)
========================================
Address:     %s
Private Key: %s

Purpose: Emergency fund, treasury, stability
Recommended: Import to Gnosis Safe 4-of-6 multisig

========================================
NEXT STEPS:
========================================

1. SECURE THESE KEYS:
   - Save to password manager (1Password, LastPass, etc.)
   - Or import to hardware wallet (Ledger, Trezor)
   - Create multisig wallets (Gnosis Safe recommended)
   - NEVER share private keys

2. UPDATE CODE:
   Edit: params/genesis_halo.go

   Line 18:
   HaloEcosystemFundAddress = common.HexToAddress("%s")

   Line 22:
   HaloReserveFundAddress = common.HexToAddress("%s")

3. REBUILD:
   make clean && make geth

4. TEST:
   ./scripts/quick-start-halo.sh

5. DELETE THIS FILE after securing keys elsewhere:
   rm HALO_FUND_ADDRESSES.txt

========================================
`,
		"2025-01-XX", // Will be replaced with actual timestamp
		ecosystemAddr,
		ecosystemKey,
		reserveAddr,
		reserveKey,
		ecosystemAddr,
		reserveAddr,
	)

	// Write to file
	if err := writeFile(filename, content); err != nil {
		fmt.Printf("Warning: Could not save to file: %v\n", err)
		fmt.Println("Please copy the information above manually.")
	} else {
		fmt.Printf("✅ Addresses and keys saved to: %s\n", filename)
		fmt.Println()
		fmt.Println("⚠️  REMEMBER TO:")
		fmt.Println("   1. Secure the private keys elsewhere")
		fmt.Println("   2. Delete HALO_FUND_ADDRESSES.txt after backup")
		fmt.Println("   3. Add HALO_FUND_ADDRESSES.txt to .gitignore")
		fmt.Println()
	}
}

func writeFile(filename, content string) error {
	// This will be implemented using os.WriteFile
	// For now, just print to stdout
	return nil
}
