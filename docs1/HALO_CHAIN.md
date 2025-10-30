# Halo Chain - Implementation Guide

## Overview

Halo is a custom EVM-compatible blockchain built on core-geth with the following specifications:

- **Chain ID**: 12000
- **Network ID**: 12000
- **Token**: HALO
- **Consensus**: Proof of Work (Ethash with custom parameters)
- **Block Time**: 1 second
- **EIP-1559**: Enabled with custom 4-way fee distribution

## Key Features

### 1. Fast Block Time
- **Target Block Time**: 1 second
- **Difficulty Adjustment**: Custom parameters for rapid blocks
- **DifficultyBoundDivisor**: 2048
- **DurationLimit**: 1 second

### 2. Custom DAG Parameters
- **Initial DAG Size**: 512 MB (vs Ethereum's 1 GB)
- **DAG Growth**: 1 MB per epoch (vs Ethereum's 8 MB)
- **Epoch Length**: 30,000 blocks
- **Purpose**: Lower memory requirements for miners

### 3. Modified Uncle System
- **MaxUncles**: 1 (vs Ethereum's 2)
- **MaxUnclesDepth**: 2 blocks
- **Uncle Rewards** (reduced to prevent self-mining attacks):
  - Depth 1: 50% of block reward (500/1000)
  - Depth 2: 37.5% of block reward (375/1000)
  - Depth 3+: 0%
- **Nephew Reward**: 1.5% of block reward per uncle (15/1000)

### 4. Block Reward Schedule
Phased reduction model with 100M max supply (at 4-second blocks):
- **Blocks 0-25,000**: 40 HALO per block (~1.16 days)
- **Blocks 25,001-358,333**: 3 HALO per block (~15.43 days)
- **Blocks 358,334-691,666**: 1.5 HALO per block (~15.43 days)
- **Blocks 691,667-5,691,666**: 1 HALO per block (~231 days)
- **Blocks 5,691,667-7,884,000**: 0.5 HALO per block (~102 days)
- **Year 2+**: 75% annual reduction, min 0.125 HALO floor

**Year 1 Total**: 8.596M HALO (8.6% of max supply)
**Maximum Supply**: 100,000,000 HALO (reached in ~90 years)

Example milestones:
```
Block          | Reward      | Phase
---------------|-------------|-------------------
0              | 40 HALO     | Early network security
25,000         | 40 HALO     | End Phase 1
25,001         | 3 HALO      | Begin Phase 2
358,333        | 3 HALO      | End Phase 2
691,666        | 1.5 HALO    | End Phase 3
5,691,666      | 1 HALO      | End Phase 4
7,884,000      | 0.5 HALO    | End Year 1
15,768,000     | 0.375 HALO  | End Year 2
```

### 5. EIP-1559 Custom Fee Distribution
Unlike standard EIP-1559 (100% burn), Halo distributes base fees as follows:

| Recipient | Percentage | Purpose |
|-----------|------------|---------|
| **Burned** | 40% | Deflationary mechanism |
| **Miners** | 30% | Mining incentive |
| **Ecosystem Fund** | 20% | Development & grants |
| **Reserve Fund** | 10% | Treasury & emergencies |

**Important**: Priority fees (tips) go 100% to miners.

### 6. Per-Contract Fee Sharing
Inspired by Sonic's model, contracts can opt-in to receive a portion of transaction fees:
- Contracts register for fee sharing
- Specify recipient address and percentage (0-100%)
- Fees from transactions interacting with the contract are shared
- **Status**: Framework implemented, storage mechanism needs completion

### 7. Gas Limits
- **Genesis Gas Limit**: 150,000,000 (150M)
- **Minimum Gas Limit**: 50,000,000 (50M)
- **Maximum Gas Limit**: 300,000,000 (300M)
- **GasLimitBoundDivisor**: 2048

### 8. EIP-1559 Parameters
- **InitialBaseFee**: 1 Gwei (1,000,000,000 wei)
- **BaseFeeChangeDenominator**: 8
- **ElasticityMultiplier**: 2

### 9. Performance Tuning
- **StateCacheSize**: 1,000,000 entries
- **CodeCacheSize**: 100,000 entries
- **TrieCleanCacheSize**: 512 MB
- **TrieDirtyCacheSize**: 256 MB
- **DatabaseCache**: 2048 MB (2 GB)

### 10. Transaction Pool
- **GlobalSlots**: 8,192
- **GlobalQueue**: 4,096
- **AccountSlots**: 128
- **AccountQueue**: 64

### 11. Network Settings
- **MaxMessageSize**: 10 MB
- **MaxPeers**: 50

### 12. Database Backend
- **PebbleDB**: Configured (vs LevelDB)
- Better performance and lower memory usage

## Files Created

### Configuration Files
1. **params/config_halo.go** - Chain configuration and EIP activation
2. **params/genesis_halo.go** - Genesis block definition and fund addresses
3. **params/bootnodes_halo.go** - Bootstrap nodes configuration
4. **params/vars/halo_vars.go** - Consensus and network parameters

### Reward Logic
5. **params/mutations/rewards_halo.go** - Block reward calculation
6. **params/mutations/rewards_halo_test.go** - Reward logic tests

### EIP-1559 Implementation
7. **consensus/misc/eip1559/eip1559_halo.go** - Custom fee distribution and contract fee sharing

### Tests
8. **params/config_halo_test.go** - Configuration tests
9. **params/example_halo_test.go** - Genesis block tests

### Integration
- Modified **params/mutations/rewards.go** to detect Chain ID 12000 and use Halo rewards

## Pre-Deployment Checklist

### Critical: Update Fund Addresses
Before deploying, you **MUST** update these addresses in `params/genesis_halo.go`:

```go
// REPLACE THESE BEFORE DEPLOYMENT
HaloEcosystemFundAddress = common.HexToAddress("0x0000000000000000000000000000000000000001")
HaloReserveFundAddress = common.HexToAddress("0x0000000000000000000000000000000000000002")
```

⚠️ **WARNING**: Using zero addresses will cause the EIP-1559 fee distribution to fail!

### Set Genesis Timestamp
Update in `params/genesis_halo.go`:
```go
Timestamp: 1700000000, // TODO: Set to actual launch timestamp
```

### Configure Bootnodes
After setting up initial nodes, update `params/bootnodes_halo.go` with real enode URLs.

## Deployment Steps

### 1. Build geth
```bash
make geth
```

### 2. Create Genesis JSON
```bash
./build/bin/geth --datadir=./halo-data init \
  <(cat <<EOF
{
  "config": {
    "networkId": 12000,
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
    "eip1559Block": 0,
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
  "alloc": {
    "0000000000000000000000000000000000000001": {
      "balance": "0x0"
    },
    "0000000000000000000000000000000000000002": {
      "balance": "0x0"
    }
  },
  "number": "0x0",
  "gasUsed": "0x0",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
EOF
)
```

### 3. Get Genesis Hash
After initialization, get the genesis hash:
```bash
./build/bin/geth --datadir=./halo-data --exec 'eth.getBlock(0).hash' console
```

Update `HaloGenesisHash` in `params/genesis_halo.go` with this value.

### 4. Start First Node (Bootnode)
```bash
./build/bin/geth \
  --datadir=./halo-data \
  --networkid 12000 \
  --port 30303 \
  --http \
  --http.api eth,net,web3,personal,miner \
  --http.port 8545 \
  --http.addr 0.0.0.0 \
  --http.corsdomain "*" \
  --mine \
  --miner.threads=2 \
  --miner.etherbase=0xYOUR_MINER_ADDRESS
```

### 5. Get Bootnode Enode
```bash
./build/bin/geth --datadir=./halo-data --exec 'admin.nodeInfo.enode' attach
```

Add this to `HaloBootnodes` in `params/bootnodes_halo.go`.

### 6. Start Additional Nodes
```bash
./build/bin/geth \
  --datadir=./halo-data-node2 \
  --networkid 12000 \
  --port 30304 \
  --bootnodes="enode://YOUR_BOOTNODE_ENODE" \
  --http \
  --http.port 8546 \
  --mine \
  --miner.threads=1
```

## Testing

### Run Configuration Tests
```bash
go test -v ./params -run TestHaloChainConfig
go test -v ./params -run TestHaloGenesisAddressesNotZero
go test -v ./params -run TestDefaultHaloGenesisBlock
```

### Run Reward Tests
```bash
go test -v ./params/mutations -run TestGetHaloBlockReward
go test -v ./params/mutations -run TestGetHaloUncleReward
go test -v ./params/mutations -run TestHaloBlockReward
```

### Verify Reward Schedule
Test that rewards follow the expected schedule:
```bash
go test -v ./params/mutations -run TestGetHaloBlockReward
```

## Monitoring & Verification

### Check Block Rewards
```javascript
// In geth console
eth.getBlock(0).miner  // Should show miner address
eth.getBalance(eth.getBlock(100000).miner)  // Check accumulated rewards
```

### Verify EIP-1559 Distribution
Monitor fund addresses to ensure they receive fees:
```javascript
eth.getBalance("0xECOSYSTEM_FUND_ADDRESS")
eth.getBalance("0xRESERVE_FUND_ADDRESS")
```

### Check Uncle Inclusion
```javascript
eth.getBlock(BLOCK_NUMBER).uncles.length  // Should be 0 or 1
```

## Network Configuration

### MetaMask Configuration
Users can add Halo network with these parameters:
- **Network Name**: Halo Network
- **RPC URL**: http://YOUR_NODE_IP:8545 (or https://rpc.halo.network)
- **Chain ID**: 12000
- **Currency Symbol**: HALO
- **Block Explorer URL**: (To be set up)

## Security Considerations

1. **Fund Address Security**
   - Use multisig for ecosystem and reserve funds
   - Implement timelock for large withdrawals
   - Regular audits of fund usage

2. **Consensus Security**
   - 1-second block time requires robust network
   - Monitor uncle rate (should be low)
   - Ensure sufficient hashrate distribution

3. **EIP-1559 Validation**
   - Fund addresses cannot be zero
   - Validation occurs during block finalization
   - Monitor burn rate vs distribution

## Future Enhancements

### Short Term
1. Complete per-contract fee sharing storage mechanism
2. Set up DNS discovery for bootnodes
3. Deploy block explorer
4. Create faucet for testnet

### Medium Term
1. Implement governance for fund management
2. Create SDK/libraries for dApp developers
3. Set up monitoring dashboards
4. Implement fee sharing registry contract

### Long Term
1. Explore state pruning optimizations
2. Consider ZK-rollup compatibility
3. Evaluate cross-chain bridge solutions

## Maintenance

### Upgrading Node Software
```bash
# Stop node gracefully
kill -TERM $(pgrep geth)

# Backup data
tar -czf halo-backup-$(date +%Y%m%d).tar.gz ./halo-data

# Build new version
git pull
make geth

# Restart node
./build/bin/geth --datadir=./halo-data --networkid 12000 ...
```

### Database Maintenance
```bash
# Check database size
du -sh ./halo-data

# Prune ancient data (if needed)
geth snapshot prune-state --datadir ./halo-data
```

## Support & Resources

- **GitHub**: https://github.com/YOUR_ORG/halo-chain
- **Documentation**: https://docs.halo.network
- **Discord**: https://discord.gg/halo
- **Twitter**: @HaloNetwork

## License

Halo Chain implementation is licensed under LGPL-3.0, inheriting from core-geth.

---

**Last Updated**: 2025-01-XX
**Version**: 1.0.0
**Core-Geth Base**: Latest
