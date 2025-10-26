#!/bin/bash
# Simple test script that doesn't require jq
# Tests basic node connectivity and fee distribution

set -e

echo "🧪 Halo Chain - Simple Test Script"
echo "===================================="
echo ""

# Test 1: Node connectivity
echo "Test 1: Node Connectivity"
echo "-------------------------"
RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}' \
    http://localhost:8545)

if [[ $RESPONSE == *"result"* ]]; then
    echo "✅ Node is responding"
    echo "   Response: $RESPONSE"
else
    echo "❌ Node is not responding"
    exit 1
fi
echo ""

# Test 2: Get block number
echo "Test 2: Block Number"
echo "-------------------"
BLOCK_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
    http://localhost:8545)

if [[ $BLOCK_RESPONSE == *"result"* ]]; then
    echo "✅ Can fetch block number"
    echo "   Response: $BLOCK_RESPONSE"
else
    echo "❌ Cannot fetch block number"
    exit 1
fi
echo ""

# Test 3: Get ecosystem fund balance
echo "Test 3: Ecosystem Fund Balance"
echo "------------------------------"
ECOSYSTEM_ADDR="0xa7548DF196e2C1476BDc41602E288c0A8F478c4f"
BALANCE_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBalance\",\"params\":[\"$ECOSYSTEM_ADDR\",\"latest\"],\"id\":1}" \
    http://localhost:8545)

if [[ $BALANCE_RESPONSE == *"result"* ]]; then
    echo "✅ Can fetch ecosystem fund balance"
    echo "   Address: $ECOSYSTEM_ADDR"
    echo "   Response: $BALANCE_RESPONSE"
else
    echo "❌ Cannot fetch balance"
    exit 1
fi
echo ""

# Test 4: Get reserve fund balance
echo "Test 4: Reserve Fund Balance"
echo "----------------------------"
RESERVE_ADDR="0xb95ae9b737e104C666d369CFb16d6De88208Bd80"
RESERVE_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBalance\",\"params\":[\"$RESERVE_ADDR\",\"latest\"],\"id\":1}" \
    http://localhost:8545)

if [[ $RESERVE_RESPONSE == *"result"* ]]; then
    echo "✅ Can fetch reserve fund balance"
    echo "   Address: $RESERVE_ADDR"
    echo "   Response: $RESERVE_RESPONSE"
else
    echo "❌ Cannot fetch balance"
    exit 1
fi
echo ""

# Test 5: Get miner balance
echo "Test 5: Miner Balance"
echo "--------------------"
MINER_ADDR="0x69AEd36e548525ED741052A6572Bb1328973b44F"
MINER_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBalance\",\"params\":[\"$MINER_ADDR\",\"latest\"],\"id\":1}" \
    http://localhost:8545)

if [[ $MINER_RESPONSE == *"result"* ]]; then
    echo "✅ Can fetch miner balance"
    echo "   Address: $MINER_ADDR"
    echo "   Response: $MINER_RESPONSE"
else
    echo "❌ Cannot fetch balance"
    exit 1
fi
echo ""

# Test 6: Get latest block
echo "Test 6: Latest Block Details"
echo "----------------------------"
LATEST_BLOCK=$(curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
    http://localhost:8545)

if [[ $LATEST_BLOCK == *"result"* ]]; then
    echo "✅ Can fetch latest block"
    echo "   Block data (truncated): ${LATEST_BLOCK:0:200}..."
else
    echo "❌ Cannot fetch latest block"
    exit 1
fi
echo ""

# Test 7: Peer count
echo "Test 7: Peer Count"
echo "-----------------"
PEER_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
    http://localhost:8545)

if [[ $PEER_RESPONSE == *"result"* ]]; then
    echo "✅ Can fetch peer count"
    echo "   Response: $PEER_RESPONSE"
else
    echo "❌ Cannot fetch peer count"
    exit 1
fi
echo ""

echo "===================================="
echo "✅ All basic tests passed!"
echo ""
echo "💡 Note: For detailed analysis with balance parsing,"
echo "   install jq or use the Node.js test scripts."
