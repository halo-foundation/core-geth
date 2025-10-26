#!/bin/bash

# Halo Chain - Fund Address Setup
# This script generates addresses AND updates the configuration automatically

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "=============================================="
echo "  Halo Chain - Fund Address Setup"
echo "  (Generate + Auto-Configure)"
echo "=============================================="
echo -e "${NC}"

# Check if geth binary exists
if [ ! -f "./build/bin/geth" ]; then
    echo -e "${RED}ERROR: geth binary not found!${NC}"
    echo "Please build first: make geth"
    exit 1
fi

# Create temporary directory for keys
TEMP_DIR="./temp-fund-keys"
mkdir -p "$TEMP_DIR"

echo -e "${GREEN}[STEP 1]${NC} Generating Ecosystem Fund Address..."
echo

# Create password file
PASSWORD="HaloChain2025SecurePassword!"
echo "$PASSWORD" > "$TEMP_DIR/password.txt"

# Generate ecosystem fund account
ECOSYSTEM_OUTPUT=$(./build/bin/geth account new --datadir "$TEMP_DIR/ecosystem" --password "$TEMP_DIR/password.txt" 2>&1)
ECOSYSTEM_ADDR=$(echo "$ECOSYSTEM_OUTPUT" | grep -oP '(?<=Public address of the key:   )0x[a-fA-F0-9]{40}')

if [ -z "$ECOSYSTEM_ADDR" ]; then
    # Try alternative grep pattern
    ECOSYSTEM_ADDR=$(echo "$ECOSYSTEM_OUTPUT" | grep -oP '0x[a-fA-F0-9]{40}' | head -1)
fi

echo -e "${BLUE}Ecosystem Fund:${NC} $ECOSYSTEM_ADDR"
echo

echo -e "${GREEN}[STEP 2]${NC} Generating Reserve Fund Address..."
echo

# Generate reserve fund account
RESERVE_OUTPUT=$(./build/bin/geth account new --datadir "$TEMP_DIR/reserve" --password "$TEMP_DIR/password.txt" 2>&1)
RESERVE_ADDR=$(echo "$RESERVE_OUTPUT" | grep -oP '(?<=Public address of the key:   )0x[a-fA-F0-9]{40}')

if [ -z "$RESERVE_ADDR" ]; then
    # Try alternative grep pattern
    RESERVE_ADDR=$(echo "$RESERVE_OUTPUT" | grep -oP '0x[a-fA-F0-9]{40}' | head -1)
fi

echo -e "${BLUE}Reserve Fund:${NC}   $RESERVE_ADDR"
echo

echo -e "${GREEN}[STEP 3]${NC} Updating configuration file..."
echo

# Backup original file
cp params/genesis_halo.go params/genesis_halo.go.backup

# Update the addresses in genesis_halo.go using sed
sed -i "s/HaloEcosystemFundAddress = common.HexToAddress(\"0x[a-fA-F0-9]*\")/HaloEcosystemFundAddress = common.HexToAddress(\"$ECOSYSTEM_ADDR\")/" params/genesis_halo.go

sed -i "s/HaloReserveFundAddress = common.HexToAddress(\"0x[a-fA-F0-9]*\")/HaloReserveFundAddress = common.HexToAddress(\"$RESERVE_ADDR\")/" params/genesis_halo.go

echo -e "${GREEN}✓${NC} Configuration updated successfully!"
echo

# Save comprehensive information to file
OUTPUT_FILE="HALO_FUND_ADDRESSES_SECURE.txt"

cat > "$OUTPUT_FILE" <<EOF
============================================
HALO CHAIN - FUND ADDRESSES
============================================
Generated: $(date)
Script: setup-fund-addresses.sh

⚠️  CRITICAL SECURITY INFORMATION ⚠️
This file contains sensitive information.
- DO NOT commit to git
- DO NOT share publicly
- SECURE immediately after reading

============================================
ECOSYSTEM FUND (20% of base fees)
============================================
Address:  $ECOSYSTEM_ADDR
Keystore: $TEMP_DIR/ecosystem/keystore/
Password: $PASSWORD

Status: ✅ CONFIGURED in params/genesis_halo.go

Purpose: Development, grants, marketing
Recommended: Transfer to Gnosis Safe 3-of-5 multisig

============================================
RESERVE FUND (10% of base fees)
============================================
Address:  $RESERVE_ADDR
Keystore: $TEMP_DIR/reserve/keystore/
Password: $PASSWORD

Status: ✅ CONFIGURED in params/genesis_halo.go

Purpose: Emergency fund, treasury, stability
Recommended: Transfer to Gnosis Safe 4-of-6 multisig

============================================
EXPORTING PRIVATE KEYS
============================================

⚠️  ONLY export private keys in a SECURE environment!

Ecosystem Fund:
./build/bin/geth account export $ECOSYSTEM_ADDR --datadir $TEMP_DIR/ecosystem
(Password: $PASSWORD)

Reserve Fund:
./build/bin/geth account export $RESERVE_ADDR --datadir $TEMP_DIR/reserve
(Password: $PASSWORD)

============================================
CONFIGURATION STATUS
============================================

✅ Generated: Ecosystem Fund address
✅ Generated: Reserve Fund address
✅ Updated:   params/genesis_halo.go (line 18)
✅ Updated:   params/genesis_halo.go (line 22)
✅ Backup:    params/genesis_halo.go.backup

NEXT STEP: Rebuild the project!
   make clean && make geth

============================================
SECURITY CHECKLIST
============================================

FOR TESTNET (Current):
- [✓] Addresses generated
- [✓] Configuration updated
- [ ] Rebuild project
- [ ] Test network deployment
- [ ] Verify fee distribution

FOR MAINNET (Future):
- [ ] Export private keys to secure storage
- [ ] Create Gnosis Safe multisigs
- [ ] Transfer funds to multisigs
- [ ] Update configuration with multisig addresses
- [ ] Delete test keystores
- [ ] Security audit
- [ ] Bug bounty program

============================================
CLEANUP AFTER SECURING KEYS
============================================

After you have securely backed up the keys:

1. Export and save private keys
2. Import to hardware wallet or multisig
3. Delete temporary files:
   rm -rf $TEMP_DIR
   rm $OUTPUT_FILE
   rm params/genesis_halo.go.backup

⚠️  DO NOT delete until keys are secured elsewhere!

============================================
TROUBLESHOOTING
============================================

If configuration update failed:
1. Check params/genesis_halo.go manually
2. Restore backup if needed:
   cp params/genesis_halo.go.backup params/genesis_halo.go
3. Update manually:
   Line 18: HaloEcosystemFundAddress = common.HexToAddress("$ECOSYSTEM_ADDR")
   Line 22: HaloReserveFundAddress = common.HexToAddress("$RESERVE_ADDR")

============================================
EOF

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  SETUP COMPLETE!${NC}"
echo -e "${GREEN}============================================${NC}"
echo
echo -e "${BLUE}Generated Addresses:${NC}"
echo "  Ecosystem: $ECOSYSTEM_ADDR"
echo "  Reserve:   $RESERVE_ADDR"
echo
echo -e "${GREEN}Configuration Updated:${NC}"
echo "  ✓ params/genesis_halo.go"
echo "  ✓ Backup saved: params/genesis_halo.go.backup"
echo
echo -e "${BLUE}Information Saved:${NC}"
echo "  📄 $OUTPUT_FILE"
echo "  🔑 Keystores: $TEMP_DIR/"
echo "  🔒 Password: $PASSWORD"
echo
echo -e "${YELLOW}⚡ NEXT STEPS:${NC}"
echo "  1. Read: $OUTPUT_FILE"
echo "  2. Rebuild: make clean && make geth"
echo "  3. Test: ./scripts/quick-start-halo.sh"
echo
echo -e "${RED}⚠️  SECURITY REMINDER:${NC}"
echo "  - These are TESTNET addresses"
echo "  - For MAINNET: Use hardware wallets + multisig"
echo "  - Secure private keys before deploying"
echo "  - Delete temp files after backup"
echo

