#!/bin/bash
# Verify Halo Chain block structure implementation
# Checks: 1-second block time, MaxUncles=1, fee distribution

set -e

echo "🔍 Verifying Halo Chain Block Structure"
echo "========================================"
echo ""

# Function to get block by number
get_block() {
    local block_num=$1
    curl -s -X POST -H "Content-Type: application/json" \
        --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$block_num\",true],\"id\":1}" \
        http://localhost:8545 | jq '.'
}

# Get latest block
echo "📦 Fetching latest blocks..."
LATEST=$(curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
    http://localhost:8545 | jq -r '.result')

LATEST_DEC=$((16#${LATEST#0x}))
echo "   Latest block: $LATEST_DEC"
echo ""

# Fetch last 10 blocks
echo "📊 Analyzing last 10 blocks..."
echo ""

PREV_TIMESTAMP=0
BLOCK_TIME_SUM=0
UNCLE_COUNT=0
BLOCKS_CHECKED=0

for i in {0..9}; do
    BLOCK_NUM=$((LATEST_DEC - i))
    BLOCK_HEX=$(printf "0x%x" $BLOCK_NUM)

    BLOCK_DATA=$(get_block "$BLOCK_HEX")

    TIMESTAMP=$(echo "$BLOCK_DATA" | jq -r '.result.timestamp')
    TIMESTAMP_DEC=$((16#${TIMESTAMP#0x}))

    UNCLES=$(echo "$BLOCK_DATA" | jq -r '.result.uncles | length')
    GAS_USED=$(echo "$BLOCK_DATA" | jq -r '.result.gasUsed')
    MINER=$(echo "$BLOCK_DATA" | jq -r '.result.miner')
    TX_COUNT=$(echo "$BLOCK_DATA" | jq -r '.result.transactions | length')

    # Calculate block time
    if [ $PREV_TIMESTAMP -ne 0 ]; then
        BLOCK_TIME=$((PREV_TIMESTAMP - TIMESTAMP_DEC))
        BLOCK_TIME_SUM=$((BLOCK_TIME_SUM + BLOCK_TIME))
        BLOCKS_CHECKED=$((BLOCKS_CHECKED + 1))

        echo "Block #$BLOCK_NUM:"
        echo "  Time: $BLOCK_TIME seconds"
        echo "  Uncles: $UNCLES"
        echo "  Transactions: $TX_COUNT"
        echo "  Gas Used: $GAS_USED"
        echo "  Miner: $MINER"
        echo ""
    fi

    PREV_TIMESTAMP=$TIMESTAMP_DEC
    UNCLE_COUNT=$((UNCLE_COUNT + UNCLES))
done

# Calculate average block time
if [ $BLOCKS_CHECKED -gt 0 ]; then
    AVG_BLOCK_TIME=$((BLOCK_TIME_SUM / BLOCKS_CHECKED))
    echo "📈 Statistics:"
    echo "   Average block time: $AVG_BLOCK_TIME seconds (target: 1 second)"
    echo "   Total uncles in last 10 blocks: $UNCLE_COUNT"
    echo ""

    # Verify block time
    if [ $AVG_BLOCK_TIME -le 2 ]; then
        echo "   ✅ Block time is correct (~1 second)"
    else
        echo "   ⚠️  Block time is higher than expected"
    fi

    # Check uncles
    if [ $UNCLE_COUNT -le 10 ]; then
        echo "   ✅ Uncle count is reasonable (MaxUncles=1 allows max 10 uncles in 10 blocks)"
    else
        echo "   ⚠️  High uncle count detected"
    fi
fi

echo ""
echo "🔍 Checking block structure in code..."
echo ""

# Check MaxUncles in consensus
if grep -q "maxUncles.*1" consensus/ethash/consensus.go 2>/dev/null; then
    echo "   ✅ MaxUncles = 1 configured"
else
    echo "   ⚠️  MaxUncles configuration not found or not set to 1"
fi

# Check block reward configuration
if grep -q "HaloBlockReward" params/vars/halo_vars.go 2>/dev/null; then
    echo "   ✅ Halo block rewards configured"
else
    echo "   ⚠️  Halo block rewards not found"
fi

# Check fee distribution
if grep -q "HaloEcosystemFundAddress\|HaloReserveFundAddress" params/genesis_halo.go 2>/dev/null; then
    echo "   ✅ Fund addresses configured"
else
    echo "   ⚠️  Fund addresses not found"
fi

echo ""
echo "========================================"
echo "✅ Block structure verification complete"
