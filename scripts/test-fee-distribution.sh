#!/bin/bash
# Test Halo Chain fee distribution
# This script tests the 4-way fee split: 40% burn, 30% miner, 20% ecosystem, 10% reserve

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "🧪 Testing Halo Chain Fee Distribution"
echo "======================================="
echo ""

# Load account information
ECOSYSTEM_ADDR="0xa7548DF196e2C1476BDc41602E288c0A8F478c4f"
RESERVE_ADDR="0xb95ae9b737e104C666d369CFb16d6De88208Bd80"
MINER_ADDR="0x69AEd36e548525ED741052A6572Bb1328973b44F"

echo "📋 Fund Addresses:"
echo "  Ecosystem: $ECOSYSTEM_ADDR (20%)"
echo "  Reserve:   $RESERVE_ADDR (10%)"
echo "  Miner:     $MINER_ADDR (30% + block rewards)"
echo ""

# Function to get balance
get_balance() {
    local addr=$1
    curl -s -X POST -H "Content-Type: application/json" \
        --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBalance\",\"params\":[\"$addr\",\"latest\"],\"id\":1}" \
        http://localhost:8545 | jq -r '.result'
}

# Function to convert hex to decimal
hex_to_dec() {
    echo $((16#${1#0x}))
}

# Function to get block number
get_block_number() {
    curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
        http://localhost:8545 | jq -r '.result'
}

# Function to get block details
get_block() {
    local block_num=$1
    curl -s -X POST -H "Content-Type: application/json" \
        --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$block_num\",true],\"id\":1}" \
        http://localhost:8545 | jq '.'
}

echo "📊 Checking initial balances..."
ECOSYSTEM_BALANCE_BEFORE=$(get_balance $ECOSYSTEM_ADDR)
RESERVE_BALANCE_BEFORE=$(get_balance $RESERVE_ADDR)
MINER_BALANCE_BEFORE=$(get_balance $MINER_ADDR)

echo "  Ecosystem: $ECOSYSTEM_BALANCE_BEFORE"
echo "  Reserve:   $RESERVE_BALANCE_BEFORE"
echo "  Miner:     $MINER_BALANCE_BEFORE"
echo ""

# Get current block number
BLOCK_NUM=$(get_block_number)
BLOCK_DEC=$(hex_to_dec $BLOCK_NUM)
echo "📦 Current block: $BLOCK_DEC ($BLOCK_NUM)"
echo ""

# Wait for a few blocks to be mined
echo "⏳ Waiting for 5 blocks to be mined..."
TARGET_BLOCK=$((BLOCK_DEC + 5))
while [ $(hex_to_dec $(get_block_number)) -lt $TARGET_BLOCK ]; do
    sleep 1
    echo -n "."
done
echo ""
echo ""

echo "📊 Checking balances after mining..."
ECOSYSTEM_BALANCE_AFTER=$(get_balance $ECOSYSTEM_ADDR)
RESERVE_BALANCE_AFTER=$(get_balance $RESERVE_ADDR)
MINER_BALANCE_AFTER=$(get_balance $MINER_ADDR)

echo "  Ecosystem: $ECOSYSTEM_BALANCE_AFTER"
echo "  Reserve:   $RESERVE_BALANCE_AFTER"
echo "  Miner:     $MINER_BALANCE_AFTER"
echo ""

# Calculate differences
ECOSYSTEM_DIFF=$(($(hex_to_dec $ECOSYSTEM_BALANCE_AFTER) - $(hex_to_dec $ECOSYSTEM_BALANCE_BEFORE)))
RESERVE_DIFF=$(($(hex_to_dec $RESERVE_BALANCE_AFTER) - $(hex_to_dec $RESERVE_BALANCE_BEFORE)))
MINER_DIFF=$(($(hex_to_dec $MINER_BALANCE_AFTER) - $(hex_to_dec $MINER_BALANCE_BEFORE)))

echo "💰 Balance changes:"
echo "  Ecosystem: +$ECOSYSTEM_DIFF wei"
echo "  Reserve:   +$RESERVE_DIFF wei"
echo "  Miner:     +$MINER_DIFF wei"
echo ""

# Check if balances increased
if [ $ECOSYSTEM_DIFF -gt 0 ]; then
    echo "✅ Ecosystem fund received fees"
else
    echo "❌ Ecosystem fund did not receive fees"
fi

if [ $RESERVE_DIFF -gt 0 ]; then
    echo "✅ Reserve fund received fees"
else
    echo "❌ Reserve fund did not receive fees"
fi

if [ $MINER_DIFF -gt 0 ]; then
    echo "✅ Miner received rewards"
else
    echo "❌ Miner did not receive rewards"
fi

echo ""
echo "📊 Ratio Analysis:"
if [ $ECOSYSTEM_DIFF -gt 0 ] && [ $RESERVE_DIFF -gt 0 ]; then
    RATIO=$(echo "scale=2; $ECOSYSTEM_DIFF / $RESERVE_DIFF" | bc)
    echo "  Ecosystem/Reserve ratio: $RATIO (expected: 2.0 for 20%/10%)"

    # Check if ratio is approximately 2.0 (within 10% margin)
    if (( $(echo "$RATIO > 1.8 && $RATIO < 2.2" | bc -l) )); then
        echo "  ✅ Ratio is correct (within margin)"
    else
        echo "  ⚠️  Ratio deviation detected"
    fi
fi

echo ""
echo "🔍 Recent block details:"
LATEST_BLOCK=$(get_block "latest")
echo "$LATEST_BLOCK" | jq '{number, miner, gasUsed, baseFeePerGas, transactions: (.transactions | length)}'

echo ""
echo "================================"
echo "✅ Fee distribution test complete"
