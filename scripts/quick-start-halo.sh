#!/bin/bash

# Halo Chain - Quick Start Script
# This script helps you quickly set up a local Halo testnet

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CHAIN_ID=12000
NETWORK_ID=12000
DATA_DIR="./halo-testnet"
GENESIS_FILE="./halo_genesis.json"

# Functions
print_header() {
    echo -e "${BLUE}"
    echo "================================================"
    echo "  Halo Chain - Quick Start Script"
    echo "  Chain ID: ${CHAIN_ID}"
    echo "================================================"
    echo -e "${NC}"
}

print_step() {
    echo -e "${GREEN}[STEP]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

check_dependencies() {
    print_step "Checking dependencies..."

    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or higher."
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Go version: ${GO_VERSION}"

    if ! command -v make &> /dev/null; then
        print_error "Make is not installed. Please install make."
        exit 1
    fi

    print_success "All dependencies found"
}

build_geth() {
    print_step "Building geth..."

    if [ ! -f "./Makefile" ]; then
        print_error "Makefile not found. Please run this script from the core-geth root directory."
        exit 1
    fi

    make geth

    if [ ! -f "./build/bin/geth" ]; then
        print_error "Build failed. geth binary not found."
        exit 1
    fi

    print_success "Geth built successfully"
    ./build/bin/geth version
}

create_genesis() {
    print_step "Creating genesis file..."

    if [ -f "${GENESIS_FILE}" ]; then
        print_warning "Genesis file already exists. Using existing file."
        return
    fi

    cat > "${GENESIS_FILE}" <<'EOF'
{
  "config": {
    "chainId": 12000,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "shanghaiBlock": 0,
    "ethash": {}
  },
  "nonce": "0x0",
  "timestamp": "0x65700000",
  "extraData": "0x48616c6f204e6574776f726b",
  "gasLimit": "0x8F0D180",
  "difficulty": "0x20000",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "coinbase": "0x0000000000000000000000000000000000000000",
  "alloc": {},
  "number": "0x0",
  "gasUsed": "0x0",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
EOF

    print_success "Genesis file created: ${GENESIS_FILE}"
}

initialize_node() {
    print_step "Initializing node..."

    if [ -d "${DATA_DIR}/geth" ]; then
        print_warning "Data directory already exists. Skipping initialization."
        print_warning "To reinitialize, delete ${DATA_DIR} first."
        return
    fi

    ./build/bin/geth init "${GENESIS_FILE}" --datadir "${DATA_DIR}"

    print_success "Node initialized"
}

create_account() {
    print_step "Creating mining account..."

    # Check if account already exists
    if [ -d "${DATA_DIR}/keystore" ] && [ "$(ls -A ${DATA_DIR}/keystore)" ]; then
        print_warning "Accounts already exist. Skipping account creation."
        MINER_ADDRESS=$(./build/bin/geth account list --datadir "${DATA_DIR}" | head -1 | grep -oP '0x[a-fA-F0-9]{40}')
        print_info "Using existing account: ${MINER_ADDRESS}"
        return
    fi

    # Create new account with default password
    echo "password" > "${DATA_DIR}/password.txt"
    MINER_ADDRESS=$(./build/bin/geth account new --datadir "${DATA_DIR}" --password "${DATA_DIR}/password.txt" | grep -oP '0x[a-fA-F0-9]{40}')

    print_success "Account created: ${MINER_ADDRESS}"
    print_info "Password saved to: ${DATA_DIR}/password.txt"
}

create_start_script() {
    print_step "Creating start script..."

    cat > "./start-halo-node.sh" <<EOF
#!/bin/bash

# Start Halo node
./build/bin/geth \\
  --datadir ${DATA_DIR} \\
  --networkid ${NETWORK_ID} \\
  --http \\
  --http.addr 0.0.0.0 \\
  --http.port 8545 \\
  --http.api eth,net,web3,personal,miner,admin,debug \\
  --http.corsdomain "*" \\
  --ws \\
  --ws.addr 0.0.0.0 \\
  --ws.port 8546 \\
  --ws.api eth,net,web3 \\
  --ws.origins "*" \\
  --mine \\
  --miner.threads 1 \\
  --miner.etherbase ${MINER_ADDRESS} \\
  --miner.gaslimit 150000000 \\
  --allow-insecure-unlock \\
  --verbosity 3 \\
  console
EOF

    chmod +x ./start-halo-node.sh

    print_success "Start script created: ./start-halo-node.sh"
}

create_attach_script() {
    print_step "Creating attach script..."

    cat > "./attach-halo-console.sh" <<EOF
#!/bin/bash

# Attach to running Halo node
./build/bin/geth attach http://localhost:8545
EOF

    chmod +x ./attach-halo-console.sh

    print_success "Attach script created: ./attach-halo-console.sh"
}

create_test_script() {
    print_step "Creating test script..."

    cat > "./test-halo-network.sh" <<EOF
#!/bin/bash

# Test Halo network

echo "Testing Halo network..."

# Test 1: Check if node is running
echo -n "Test 1: Node connectivity... "
if curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' http://localhost:8545 > /dev/null; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
    exit 1
fi

# Test 2: Check chain ID
echo -n "Test 2: Chain ID... "
CHAIN_ID=\$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' http://localhost:8545 | grep -oP '0x[0-9a-f]+' | head -1)
if [ "\$CHAIN_ID" = "0x2ee0" ]; then  # 12000 in hex
    echo "✅ PASS (Chain ID: \$CHAIN_ID)"
else
    echo "❌ FAIL (Expected 0x2ee0, got \$CHAIN_ID)"
fi

# Test 3: Check if mining
echo -n "Test 3: Mining status... "
MINING=\$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' http://localhost:8545 | grep -oP 'true|false')
if [ "\$MINING" = "true" ]; then
    echo "✅ PASS (Mining active)"
else
    echo "⚠️  WARN (Mining inactive)"
fi

# Test 4: Check block production
echo -n "Test 4: Block production... "
BLOCK1=\$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545 | grep -oP '0x[0-9a-f]+' | head -1)
sleep 5
BLOCK2=\$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545 | grep -oP '0x[0-9a-f]+' | head -1)
if [ "\$BLOCK2" != "\$BLOCK1" ]; then
    echo "✅ PASS (Blocks: \$BLOCK1 → \$BLOCK2)"
else
    echo "❌ FAIL (No new blocks)"
fi

# Test 5: Check account balance
echo -n "Test 5: Mining rewards... "
BALANCE=\$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["${MINER_ADDRESS}","latest"],"id":1}' http://localhost:8545 | grep -oP '0x[0-9a-f]+' | tail -1)
if [ "\$BALANCE" != "0x0" ]; then
    echo "✅ PASS (Balance: \$BALANCE)"
else
    echo "⚠️  WARN (No rewards yet)"
fi

echo ""
echo "Test complete! 🎉"
EOF

    chmod +x ./test-halo-network.sh

    print_success "Test script created: ./test-halo-network.sh"
}

print_instructions() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Setup Complete! 🎉${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "Your Halo testnet is ready!"
    echo ""
    echo -e "${BLUE}Mining Account:${NC} ${MINER_ADDRESS}"
    echo -e "${BLUE}Data Directory:${NC} ${DATA_DIR}"
    echo -e "${BLUE}Password File:${NC} ${DATA_DIR}/password.txt"
    echo ""
    echo -e "${YELLOW}Next Steps:${NC}"
    echo ""
    echo "  1. Start the node:"
    echo -e "     ${GREEN}./start-halo-node.sh${NC}"
    echo ""
    echo "  2. In the console, check mining:"
    echo -e "     ${GREEN}> eth.mining${NC}"
    echo -e "     ${GREEN}> eth.blockNumber${NC}"
    echo -e "     ${GREEN}> eth.getBalance(eth.coinbase)${NC}"
    echo ""
    echo "  3. Or attach from another terminal:"
    echo -e "     ${GREEN}./attach-halo-console.sh${NC}"
    echo ""
    echo "  4. Run tests (in another terminal):"
    echo -e "     ${GREEN}./test-halo-network.sh${NC}"
    echo ""
    echo -e "${YELLOW}Useful Commands in Console:${NC}"
    echo "  - eth.accounts              # List accounts"
    echo "  - eth.blockNumber           # Current block"
    echo "  - eth.getBalance(eth.coinbase) # Your balance"
    echo "  - miner.start(1)            # Start mining"
    echo "  - miner.stop()              # Stop mining"
    echo "  - admin.peers               # Connected peers"
    echo ""
    echo -e "${YELLOW}Connect MetaMask:${NC}"
    echo "  Network Name: Halo Local"
    echo "  RPC URL: http://localhost:8545"
    echo "  Chain ID: 12000"
    echo "  Currency Symbol: HALO"
    echo ""
    echo -e "${YELLOW}Documentation:${NC}"
    echo "  - Quick Start: HALO_QUICK_START.md"
    echo "  - Launch Guide: HALO_LAUNCH_GUIDE.md"
    echo "  - Parameters: HALO_PARAMETERS.md"
    echo ""
    echo "Happy mining! 🚀"
    echo ""
}

# Main execution
main() {
    print_header

    check_dependencies
    build_geth
    create_genesis
    initialize_node
    create_account
    create_start_script
    create_attach_script
    create_test_script

    print_instructions
}

# Run main function
main
