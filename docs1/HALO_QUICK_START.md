# Halo Chain - Quick Start Guide

**Get your Halo node running in 15 minutes**

---

## Prerequisites

- Linux/macOS/Windows with WSL2
- 8GB+ RAM
- 100GB+ free disk space
- Go 1.21+ (for building from source)

---

## Option 1: Quick Test (Single Node)

### Step 1: Update Configuration (2 minutes)

```bash
cd /home/blackluv/core-geth

# Edit fund addresses (use your own or leave as test addresses)
nano params/genesis_halo.go

# Update lines 18 and 22 with your addresses:
# HaloEcosystemFundAddress = common.HexToAddress("0xYOUR_ADDRESS")
# HaloReserveFundAddress = common.HexToAddress("0xYOUR_ADDRESS")
```

### Step 2: Build (3 minutes)

```bash
make geth

# Verify
./build/bin/geth version
```

### Step 3: Create Genesis (1 minute)

```bash
cat > halo_genesis.json <<'EOF'
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
```

### Step 4: Initialize (1 minute)

```bash
./build/bin/geth init halo_genesis.json --datadir ./halo-test
```

### Step 5: Create Mining Account (1 minute)

```bash
./build/bin/geth account new --datadir ./halo-test

# Enter a password and record the address shown
# Example: Public address of the key: 0x1234...
```

### Step 6: Start Mining! (1 second)

```bash
./build/bin/geth \
  --datadir ./halo-test \
  --networkid 12000 \
  --http \
  --http.api eth,net,web3,personal,miner \
  --http.corsdomain "*" \
  --mine \
  --miner.threads 1 \
  --miner.etherbase YOUR_MINER_ADDRESS \
  --allow-insecure-unlock \
  console
```

### Step 7: Verify It's Working

In the geth console:

```javascript
// Check if mining
> eth.mining
true

// Check block number (should increase)
> eth.blockNumber
42

// Wait 5 seconds and check again
> eth.blockNumber
47

// Check your balance (should be growing!)
> eth.getBalance(eth.coinbase)
235000000000000000000  // 235 HALO!

// Check block time (should be ~1 second)
> eth.getBlock(eth.blockNumber).timestamp - eth.getBlock(eth.blockNumber-1).timestamp
1
```

**🎉 Congratulations! You're mining HALO!**

---

## Option 2: Multi-Node Network (30 minutes)

### Terminal 1: Bootnode

```bash
# Generate bootnode key
./build/bin/bootnode -genkey bootnode.key

# Start bootnode
./build/bin/bootnode -nodekey bootnode.key -addr :30301

# Copy the enode URL shown
# enode://abc123...@127.0.0.1:30301
```

### Terminal 2: Node 1 (Miner)

```bash
# Initialize
./build/bin/geth init halo_genesis.json --datadir ./node1

# Create account
./build/bin/geth account new --datadir ./node1

# Start node
./build/bin/geth \
  --datadir ./node1 \
  --networkid 12000 \
  --port 30303 \
  --bootnodes "enode://YOUR_BOOTNODE_ENODE_FROM_TERMINAL_1" \
  --http \
  --http.port 8545 \
  --mine \
  --miner.threads 2 \
  --miner.etherbase YOUR_MINER_ADDRESS \
  console
```

### Terminal 3: Node 2 (Miner)

```bash
# Initialize
./build/bin/geth init halo_genesis.json --datadir ./node2

# Create account
./build/bin/geth account new --datadir ./node2

# Start node
./build/bin/geth \
  --datadir ./node2 \
  --networkid 12000 \
  --port 30304 \
  --bootnodes "enode://YOUR_BOOTNODE_ENODE" \
  --http \
  --http.port 8546 \
  --mine \
  --miner.threads 2 \
  --miner.etherbase YOUR_MINER_ADDRESS_2 \
  console
```

### Terminal 4: RPC Node (No mining)

```bash
# Initialize
./build/bin/geth init halo_genesis.json --datadir ./rpc-node

# Start RPC-only node
./build/bin/geth \
  --datadir ./rpc-node \
  --networkid 12000 \
  --port 30305 \
  --bootnodes "enode://YOUR_BOOTNODE_ENODE" \
  --http \
  --http.port 8547 \
  --http.api eth,net,web3,txpool \
  --ws \
  --ws.port 8548 \
  --ws.api eth,net,web3 \
  console
```

### Verify Network

In any node console:

```javascript
// Check peers (should be 2-3)
> admin.peers.length
3

// Check peer details
> admin.peers

// Check if all nodes on same block
> eth.blockNumber
// Compare across all nodes - should be same or within 1-2 blocks

// Check if uncle rate is low
> eth.getBlock("latest").uncles.length
0  // Good! Should be 0 most of the time
```

---

## Option 3: Connect to Existing Network

### If testnet/mainnet is already running:

```bash
# Download official genesis.json
wget https://raw.githubusercontent.com/YOUR_ORG/halo-chain/main/genesis.json

# Initialize
./build/bin/geth init genesis.json --datadir ./halo-node

# Start syncing
./build/bin/geth \
  --datadir ./halo-node \
  --networkid 12000 \
  --bootnodes "enode://OFFICIAL_BOOTNODE_1,enode://OFFICIAL_BOOTNODE_2" \
  --http \
  --http.api eth,net,web3 \
  console

# In console, check sync status
> eth.syncing
{
  currentBlock: 12345,
  highestBlock: 54321,
  startingBlock: 0
}

# Or if fully synced:
> eth.syncing
false
```

---

## Testing Transactions

### Send a transaction:

```javascript
// In geth console

// Unlock your account
> personal.unlockAccount(eth.accounts[0], "YOUR_PASSWORD", 300)
true

// Send HALO to another address
> eth.sendTransaction({
    from: eth.accounts[0],
    to: "0xRECIPIENT_ADDRESS",
    value: web3.toWei(10, "ether")  // 10 HALO
  })
"0xTRANSACTION_HASH..."

// Check transaction
> eth.getTransaction("0xTRANSACTION_HASH...")

// Wait for confirmation (1-2 blocks)
> eth.getTransactionReceipt("0xTRANSACTION_HASH...")
```

### Deploy a smart contract:

```javascript
// Simple contract: store a number

var contractCode = "0x608060405234801561001057600080fd5b5060c78061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060325760003560e01c80632e64cec11460375780636057361d146051575b600080fd5b603d6069565b6040516048919060a2565b60405180910390f35b6067600480360381019060639190606f565b6072565b005b60008054905090565b8060008190555050565b60008135905060898160ad565b92915050565b609c8160bb565b82525050565b600060208201905060b56000830184608f565b92915050565b6000819050919050565b60c48160bb565b811460ce57600080fd5b5056fea2646970667358221220c8f8c8f8c8f8c8f8c8f8c8f8c8f8c8f8c8f8c8f8c8f8c8f8c8f8c8f8c8f8c864736f6c63430008070033"

// Deploy
> var tx = eth.sendTransaction({
    from: eth.accounts[0],
    data: contractCode,
    gas: 200000
  })

// Get contract address
> var receipt = eth.getTransactionReceipt(tx)
> var contractAddress = receipt.contractAddress
> contractAddress
"0xCONTRACT_ADDRESS..."

// Interact with contract
> var contract = eth.contract(ABI).at(contractAddress)
> contract.store(42)  // Store number
> contract.retrieve()  // Get number
42
```

---

## Connect MetaMask

1. Open MetaMask
2. Click network dropdown
3. Click "Add Network"
4. Enter details:

```
Network Name: Halo Local
RPC URL: http://localhost:8545
Chain ID: 12000
Currency Symbol: HALO
Block Explorer: (leave blank for local)
```

5. Click "Save"
6. Switch to Halo Local network
7. Import account using private key from geth

---

## Useful Console Commands

### Account Management

```javascript
// List accounts
> eth.accounts

// Create new account
> personal.newAccount("password")

// Check balance
> eth.getBalance(eth.accounts[0])
> web3.fromWei(eth.getBalance(eth.accounts[0]), "ether")

// Unlock account
> personal.unlockAccount(eth.accounts[0], "password", 300)
```

### Mining

```javascript
// Start mining
> miner.start(1)  // 1 thread

// Stop mining
> miner.stop()

// Check hashrate
> miner.hashrate

// Set etherbase (where rewards go)
> miner.setEtherbase(eth.accounts[1])
```

### Network

```javascript
// Get peer count
> net.peerCount

// Get peer details
> admin.peers

// Add peer manually
> admin.addPeer("enode://...")

// Get node info
> admin.nodeInfo
```

### Blockchain

```javascript
// Get latest block
> eth.getBlock("latest")

// Get specific block
> eth.getBlock(1234)

// Get block by hash
> eth.getBlock("0xBLOCK_HASH...")

// Get transaction
> eth.getTransaction("0xTX_HASH...")

// Get receipt
> eth.getTransactionReceipt("0xTX_HASH...")
```

### Gas and Fees

```javascript
// Get current gas price
> eth.gasPrice

// Get base fee (EIP-1559)
> eth.getBlock("latest").baseFeePerGas

// Estimate gas
> eth.estimateGas({
    from: eth.accounts[0],
    to: "0xTO_ADDRESS",
    value: web3.toWei(1, "ether")
  })
```

### Fund Addresses (Halo-specific)

```javascript
// Check ecosystem fund balance
> eth.getBalance("ECOSYSTEM_FUND_ADDRESS")

// Check reserve fund balance
> eth.getBalance("RESERVE_FUND_ADDRESS")

// Verify fee distribution working
// Send transaction, wait for block, check balances increased
```

---

## Monitoring & Debugging

### Watch logs in real-time

```bash
# In separate terminal
tail -f ./halo-test/geth.log
```

### Check disk usage

```bash
du -sh ./halo-test
```

### Export logs

```bash
./build/bin/geth --datadir ./halo-test export-logs logs.txt
```

### Backup blockchain

```bash
tar -czf halo-backup.tar.gz ./halo-test/geth/chaindata
```

### Reset blockchain (keep accounts)

```bash
./build/bin/geth removedb --datadir ./halo-test
./build/bin/geth init halo_genesis.json --datadir ./halo-test
```

---

## Performance Optimization

### For Mining Nodes

```bash
./build/bin/geth \
  --datadir ./halo-node \
  --cache 2048 \
  --miner.threads 4 \
  --miner.gasprice 1000000000 \
  ...
```

### For RPC Nodes

```bash
./build/bin/geth \
  --datadir ./halo-node \
  --cache 4096 \
  --http \
  --http.api eth,net,web3,txpool \
  --http.vhosts "*" \
  --ws \
  --maxpeers 100 \
  ...
```

### For Archive Nodes

```bash
./build/bin/geth \
  --datadir ./halo-archive \
  --gcmode archive \
  --cache 8192 \
  --syncmode full \
  ...
```

---

## Troubleshooting

### Nodes not connecting?

```bash
# Check firewall
sudo ufw allow 30303/tcp
sudo ufw allow 30303/udp

# Check bootnode
nc -zv 127.0.0.1 30301

# Try adding peer manually
> admin.addPeer("enode://...")
```

### Mining not working?

```javascript
// Check if mining started
> eth.mining
true

// Check hashrate
> miner.hashrate
1234567

// Start mining manually
> miner.start(2)
```

### Blocks not producing?

```javascript
// Check difficulty
> eth.getBlock("latest").difficulty

// Might be too high, wait or add more miners
```

### Sync stuck?

```bash
# Remove peers and restart
> admin.peers.forEach(function(peer) {
    admin.removePeer(peer.enode)
  })

# Restart node
```

---

## Next Steps

1. **Read full launch guide:** `HALO_LAUNCH_GUIDE.md`
2. **Review parameters:** `HALO_PARAMETERS.md`
3. **Deploy smart contracts**
4. **Set up monitoring**
5. **Join community**

---

## Quick Reference

**Chain ID:** 12000
**Network ID:** 12000
**Currency:** HALO
**Block Time:** ~1 second
**Gas Limit:** 150M
**EIP-1559:** Custom (40% burn, 30% miner, 20% ecosystem, 10% reserve)

**Default Ports:**
- P2P: 30303
- HTTP RPC: 8545
- WebSocket: 8546
- Metrics: 6060

**Data Directory Locations:**
- Linux: `~/.ethereum/halo/`
- macOS: `~/Library/Ethereum/halo/`
- Windows: `%APPDATA%\Ethereum\halo\`

---

**Need Help?**
- Documentation: `HALO_LAUNCH_GUIDE.md`
- Discord: https://discord.gg/halo
- GitHub Issues: https://github.com/YOUR_ORG/core-geth/issues

**Happy building! 🚀**
