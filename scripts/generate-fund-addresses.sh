#!/bin/bash

# Halo Chain - Fund Address Generator
# This script generates Ethereum addresses for the ecosystem and reserve funds

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "=============================================="
echo "  Halo Chain - Fund Address Generation"
echo "=============================================="
echo -e "${NC}"

# Create temporary directory for keys
TEMP_DIR="./temp-fund-keys"
mkdir -p "$TEMP_DIR"

echo -e "${GREEN}[STEP 1]${NC} Generating Ecosystem Fund Address..."
echo

# Create password file
echo "HaloChain2025!" > "$TEMP_DIR/password.txt"

# Generate ecosystem fund account
ECOSYSTEM_OUTPUT=$(./build/bin/geth account new --datadir "$TEMP_DIR/ecosystem" --password "$TEMP_DIR/password.txt" 2>&1)
ECOSYSTEM_ADDR=$(echo "$ECOSYSTEM_OUTPUT" | grep -oP '0x[a-fA-F0-9]{40}' | head -1)

echo -e "${BLUE}Ecosystem Fund Address:${NC} $ECOSYSTEM_ADDR"
echo

echo -e "${GREEN}[STEP 2]${NC} Generating Reserve Fund Address..."
echo

# Generate reserve fund account
RESERVE_OUTPUT=$(./build/bin/geth account new --datadir "$TEMP_DIR/reserve" --password "$TEMP_DIR/password.txt" 2>&1)
RESERVE_ADDR=$(echo "$RESERVE_OUTPUT" | grep -oP '0x[a-fA-F0-9]{40}' | head -1)

echo -e "${BLUE}Reserve Fund Address:${NC} $RESERVE_ADDR"
echo

# Save to file
OUTPUT_FILE="HALO_FUND_ADDRESSES.txt"

cat > "$OUTPUT_FILE" <<EOF
============================================
HALO CHAIN - FUND ADDRESSES
============================================
Generated: $(date)

⚠️  CRITICAL: KEEP THIS FILE SECURE ⚠️
DO NOT commit to git or share publicly.

============================================
ECOSYSTEM FUND (20% of base fees)
============================================
Address:  $ECOSYSTEM_ADDR
Keystore: $TEMP_DIR/ecosystem/keystore/
Password: HaloChain2025!

Purpose: Development, grants, marketing
Recommended: Import to Gnosis Safe 3-of-5 multisig

============================================
RESERVE FUND (10% of base fees)
============================================
Address:  $RESERVE_ADDR
Keystore: $TEMP_DIR/reserve/keystore/
Password: HaloChain2025!

Purpose: Emergency fund, treasury, stability
Recommended: Import to Gnosis Safe 4-of-6 multisig

============================================
PRIVATE KEYS (CRITICAL - SECURE THESE!)
============================================

To export private keys:

1. Ecosystem Fund:
   ./build/bin/geth account export $ECOSYSTEM_ADDR --datadir $TEMP_DIR/ecosystem

2. Reserve Fund:
   ./build/bin/geth account export $RESERVE_ADDR --datadir $TEMP_DIR/reserve

Password for both: HaloChain2025!

============================================
NEXT STEPS:
============================================

1. SECURE THESE KEYS:
   - Export private keys using commands above
   - Save to password manager (1Password, LastPass, etc.)
   - Or import to hardware wallet (Ledger, Trezor)
   - Create multisig wallets (Gnosis Safe recommended)

2. UPDATE CODE:
   Edit: params/genesis_halo.go

   Line 18:
   HaloEcosystemFundAddress = common.HexToAddress("$ECOSYSTEM_ADDR")

   Line 22:
   HaloReserveFundAddress = common.HexToAddress("$RESERVE_ADDR")

3. REBUILD:
   make clean && make geth

4. TEST:
   ./scripts/quick-start-halo.sh

5. CLEANUP after securing keys:
   rm -rf $TEMP_DIR
   rm $OUTPUT_FILE

============================================
EOF

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  ADDRESSES GENERATED SUCCESSFULLY!${NC}"
echo -e "${GREEN}============================================${NC}"
echo
echo -e "${BLUE}Ecosystem Fund:${NC} $ECOSYSTEM_ADDR"
echo -e "${BLUE}Reserve Fund:${NC}   $RESERVE_ADDR"
echo
echo -e "${YELLOW}IMPORTANT:${NC}"
echo "  1. Addresses saved to: $OUTPUT_FILE"
echo "  2. Keystores saved to: $TEMP_DIR/"
echo "  3. Password: HaloChain2025!"
echo
echo -e "${RED}⚠️  SECURITY WARNING:${NC}"
echo "  - These are TESTNET addresses only!"
echo "  - For MAINNET, use hardware wallets + multisig"
echo "  - NEVER commit private keys to git"
echo "  - Delete temp files after securing keys"
echo
echo -e "${YELLOW}Next Steps:${NC}"
echo "  1. Read $OUTPUT_FILE for instructions"
echo "  2. Update params/genesis_halo.go with addresses"
echo "  3. Rebuild: make clean && make geth"
echo
