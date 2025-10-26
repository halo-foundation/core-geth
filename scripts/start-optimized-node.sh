#!/bin/bash
# Halo Chain Optimized Production Node
# Includes: Archive mode for Blockscout + Performance optimizations

set -e

DATADIR="/tmp/halo-production"
GENESIS_FILE="halo_genesis.json"
MINER_ADDRESS="0x69AEd36e548525ED741052A6572Bb1328973b44F"

echo "🚀 Starting Halo Chain Optimized Node"
echo "=========================================="
echo ""

# Check if datadir exists
if [ ! -d "$DATADIR/geth" ]; then
    echo "📁 Initializing new datadir: $DATADIR"
    ./build/bin/geth --datadir $DATADIR init $GENESIS_FILE
    echo "✅ Initialization complete"
    echo ""
fi

echo "⚙️  Configuration:"
echo "   - Datadir: $DATADIR"
echo "   - Network ID: 12000"
echo "   - Miner: $MINER_ADDRESS"
echo "   - Mode: Archive (for Blockscout)"
echo "   - Cache: 8GB"
echo "   - HTTP RPC: http://0.0.0.0:8545"
echo "   - WS RPC: ws://0.0.0.0:8546"
echo ""

echo "🔧 Optimizations:"
echo "   ✅ Archive mode (no pruning)"
echo "   ✅ 8GB cache"
echo "   ✅ 8192 txpool slots"
echo "   ✅ Trie cache optimization"
echo ""

echo "▶️  Starting node..."
echo ""

./build/bin/geth \
  --datadir $DATADIR \
  --networkid 12000 \
  \
  --http \
  --http.addr 0.0.0.0 \
  --http.port 8545 \
  --http.api eth,net,web3,personal,miner,debug,txpool \
  --http.corsdomain "*" \
  --http.vhosts "*" \
  \
  --ws \
  --ws.addr 0.0.0.0 \
  --ws.port 8546 \
  --ws.api eth,net,web3 \
  --ws.origins "*" \
  \
  --allow-insecure-unlock \
  \
  --mine \
  --miner.threads=1 \
  --miner.etherbase=$MINER_ADDRESS \
  --miner.gasprice 1000000000 \
  \
  --nodiscover \
  --maxpeers 0 \
  \
  --gcmode archive \
  --syncmode full \
  \
  --cache 8192 \
  --cache.trie 1024 \
  --cache.gc 50 \
  \
  --txpool.globalslots 8192 \
  --txpool.globalqueue 4096 \
  --txpool.accountslots 128 \
  --txpool.accountqueue 64 \
  \
  --verbosity 3 \
  --log.file $DATADIR/halo-node.log
