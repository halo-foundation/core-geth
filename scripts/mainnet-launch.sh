#!/bin/bash
# Halo Chain Mainnet Launch Script
# IMPORTANT: Only use this for mainnet launch after security audit!

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "🚀 Halo Chain - Mainnet Launch Script"
echo "======================================"
echo ""

# Safety check
echo "⚠️  WARNING: This will start the MAINNET node!"
echo ""
echo "Before proceeding, ensure you have:"
echo "  [✓] Completed security audit"
echo "  [✓] Set up multisig wallets for fund addresses"
echo "  [✓] Updated genesis with production addresses"
echo "  [✓] Configured production bootnodes"
echo "  [✓] Tested on testnet for 2+ weeks"
echo "  [✓] Removed all test keys"
echo "  [✓] Set correct genesis timestamp"
echo ""

read -p "Have you completed ALL pre-launch tasks? (yes/NO): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "❌ Launch cancelled. Please complete the checklist in MAINNET_LAUNCH_CHECKLIST.md"
    exit 1
fi

echo ""
read -p "Type 'LAUNCH MAINNET' to continue: " FINAL_CONFIRM

if [ "$FINAL_CONFIRM" != "LAUNCH MAINNET" ]; then
    echo "❌ Launch cancelled."
    exit 1
fi

echo ""
echo "🔍 Pre-flight Checks"
echo "-------------------"

# Check geth binary
if [ ! -f "$PROJECT_DIR/build/bin/geth" ]; then
    echo "❌ geth binary not found!"
    echo "   Run: make clean && make geth"
    exit 1
fi
echo "✅ geth binary found"

# Check genesis file
if [ ! -f "$PROJECT_DIR/halo_genesis.json" ]; then
    echo "❌ halo_genesis.json not found!"
    exit 1
fi
echo "✅ Genesis file found"

# Verify genesis timestamp is not placeholder
GENESIS_TIME=$(grep -o '"timestamp":"[^"]*"' "$PROJECT_DIR/halo_genesis.json" | cut -d'"' -f4)
GENESIS_DEC=$((16#${GENESIS_TIME#0x}))

if [ $GENESIS_DEC -eq 1700000000 ]; then
    echo "⚠️  WARNING: Genesis timestamp is still the default placeholder!"
    echo "   Update genesis timestamp in halo_genesis.json before mainnet launch"
    read -p "Continue anyway? (yes/NO): " TIMESTAMP_CONFIRM
    if [ "$TIMESTAMP_CONFIRM" != "yes" ]; then
        exit 1
    fi
fi
echo "✅ Genesis timestamp: $GENESIS_DEC ($(date -d @$GENESIS_DEC 2>/dev/null || echo 'N/A'))"

# Check data directory doesn't exist (fresh start)
MAINNET_DATADIR="$PROJECT_DIR/halo-mainnet"

if [ -d "$MAINNET_DATADIR" ]; then
    echo "⚠️  Mainnet data directory already exists!"
    echo "   This suggests the node has been started before."
    read -p "Continue with existing data? (yes/NO): " DATADIR_CONFIRM
    if [ "$DATADIR_CONFIRM" != "yes" ]; then
        exit 1
    fi
else
    echo "✅ Fresh mainnet data directory"
fi

echo ""
echo "⚙️  Configuration"
echo "----------------"

# Read configuration
HTTP_PORT="${HTTP_PORT:-8545}"
WS_PORT="${WS_PORT:-8546}"
P2P_PORT="${P2P_PORT:-30303}"

echo "HTTP RPC Port: $HTTP_PORT"
echo "WebSocket Port: $WS_PORT"
echo "P2P Port: $P2P_PORT"
echo ""

# Mining configuration
read -p "Enable mining on this node? (yes/NO): " ENABLE_MINING

if [ "$ENABLE_MINING" = "yes" ]; then
    read -p "Enter miner address (0x...): " MINER_ADDRESS
    if [ -z "$MINER_ADDRESS" ]; then
        echo "❌ Miner address required for mining"
        exit 1
    fi
    echo "✅ Mining will be enabled"
    echo "   Miner address: $MINER_ADDRESS"
else
    echo "ℹ️  Mining disabled"
fi

echo ""
echo "🚀 Starting Mainnet Node"
echo "------------------------"

# Create data directory
mkdir -p "$MAINNET_DATADIR"

# Initialize genesis if needed
if [ ! -d "$MAINNET_DATADIR/geth" ]; then
    echo "Initializing genesis block..."
    "$PROJECT_DIR/build/bin/geth" init \
        --datadir "$MAINNET_DATADIR" \
        "$PROJECT_DIR/halo_genesis.json"
    echo "✅ Genesis initialized"
fi

# Build geth command
GETH_CMD="$PROJECT_DIR/build/bin/geth"
GETH_ARGS=(
    "--halo"
    "--datadir" "$MAINNET_DATADIR"
    "--port" "$P2P_PORT"
    "--http"
    "--http.addr" "0.0.0.0"
    "--http.port" "$HTTP_PORT"
    "--http.api" "eth,net,web3"
    "--http.corsdomain" "*"
    "--ws"
    "--ws.addr" "0.0.0.0"
    "--ws.port" "$WS_PORT"
    "--ws.api" "eth,net,web3"
    "--ws.origins" "*"
    "--syncmode" "full"
    "--gcmode" "archive"
    "--maxpeers" "50"
)

# Add mining if enabled
if [ "$ENABLE_MINING" = "yes" ]; then
    GETH_ARGS+=("--mine")
    GETH_ARGS+=("--miner.etherbase" "$MINER_ADDRESS")
    GETH_ARGS+=("--miner.threads" "1")
fi

# Log file
LOG_FILE="$PROJECT_DIR/halo-mainnet.log"

echo ""
echo "Starting node with:"
echo "  Data directory: $MAINNET_DATADIR"
echo "  Log file: $LOG_FILE"
echo "  HTTP RPC: http://0.0.0.0:$HTTP_PORT"
echo "  WebSocket: ws://0.0.0.0:$WS_PORT"
echo "  P2P Port: $P2P_PORT"
if [ "$ENABLE_MINING" = "yes" ]; then
    echo "  Mining: ENABLED"
    echo "  Miner address: $MINER_ADDRESS"
fi
echo ""

read -p "Press ENTER to start the node..."

# Start node in background
nohup "$GETH_CMD" "${GETH_ARGS[@]}" >> "$LOG_FILE" 2>&1 &
GETH_PID=$!

echo ""
echo "✅ Mainnet node started!"
echo "   PID: $GETH_PID"
echo "   Log: tail -f $LOG_FILE"
echo ""

# Wait a moment and check if process is running
sleep 3

if ps -p $GETH_PID > /dev/null; then
    echo "✅ Node is running"
    echo ""
    echo "📊 Monitor the node:"
    echo "   tail -f $LOG_FILE"
    echo ""
    echo "🛑 Stop the node:"
    echo "   ./scripts/stop-halo.sh"
    echo "   OR: kill $GETH_PID"
    echo ""
    echo "📡 Check status:"
    echo "   curl -X POST -H 'Content-Type: application/json' \\"
    echo "     --data '{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}' \\"
    echo "     http://localhost:$HTTP_PORT"
    echo ""
    echo "🎉 MAINNET LAUNCH SUCCESSFUL!"
else
    echo "❌ Node failed to start!"
    echo "   Check logs: cat $LOG_FILE"
    exit 1
fi
