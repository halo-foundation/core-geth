# Halo Chain - Parameter Reference

Quick reference for all Halo chain custom parameters.

## Network Identity

| Parameter | Value |
|-----------|-------|
| **Chain ID** | 12000 |
| **Network ID** | 12000 |
| **Currency** | HALO |
| **Consensus** | Proof of Work (Ethash) |

## Block Parameters

| Parameter | Value | Ethereum Equivalent |
|-----------|-------|---------------------|
| **Target Block Time** | 4 seconds | 12 seconds |
| **DifficultyBoundDivisor** | 2048 | 2048 |
| **DurationLimit** | 4 seconds | 13 seconds |
| **Genesis Difficulty** | 65,536 (0x10000) | 131,072 |

## Gas Limits

| Parameter | Value | Ethereum Equivalent |
|-----------|-------|---------------------|
| **Genesis Gas Limit** | 150,000,000 | 30,000,000 |
| **Minimum Gas Limit** | 50,000,000 | 5,000 |
| **Maximum Gas Limit** | 300,000,000 | Dynamic |
| **GasLimitBoundDivisor** | 2048 | 1024 |

## Uncle Parameters

| Parameter | Value | Ethereum |
|-----------|-------|----------|
| **MaxUncles** | 1 | 2 |
| **MaxUnclesDepth** | 2 blocks | 7 blocks |
| **Uncle Reward (Depth 1)** | 50% of block reward | (uncle.number + 8 - block.number) × reward ÷ 8 |
| **Uncle Reward (Depth 2)** | 37.5% of block reward | (uncle.number + 8 - block.number) × reward ÷ 8 |
| **Nephew Reward** | 1.5% per uncle | 1/32 per uncle (3.125%) |

**Formula**:
```
Uncle Reward (Depth 1) = BlockReward × 500 / 1000
Uncle Reward (Depth 2) = BlockReward × 375 / 1000
Uncle Reward (Depth 3+) = 0
Nephew Reward = BlockReward × 15 / 1000 × uncleCount
```

**Security Note**: Reduced from 87.5%/75%/3.1% to prevent self-mining uncle attacks while still incentivizing legitimate uncle inclusion.

## Block Reward Schedule

**UPDATED for 4-second blocks** - Phased reduction schedule with 100M max supply cap:

| Block Range | Reward per Block | Duration (at 4s blocks) | Total Issuance |
|-------------|------------------|------------------------|----------------|
| 0 - 25,000 | 40 HALO | 1.16 days | 1,000,000 HALO |
| 25,001 - 358,333 | 3 HALO | 15.43 days | 999,999 HALO |
| 358,334 - 691,666 | 1.5 HALO | 15.43 days | 499,999.5 HALO |
| 691,667 - 5,691,666 | 1 HALO | 231.48 days | 5,000,000 HALO |
| 5,691,667 - 7,884,000 | 0.5 HALO | 101.50 days | 1,096,167 HALO |
| Year 2 | 0.375 HALO | 365 days | 2,956,500 HALO |
| Year 3 | 0.28125 HALO | 365 days | 2,217,375 HALO |
| Year 4 | 0.2109375 HALO | 365 days | 1,663,031 HALO |
| Year 5 | 0.158203125 HALO | 365 days | 1,247,273 HALO |
| Year 6+ | 0.125 HALO (min) | Perpetual | ~985,500 HALO/year |

**Formula**:
```javascript
function GetHaloBlockReward(blockNumber) {
  // Phase 1: blocks 0-25,000
  if (blockNumber < 25000) return 40e18;

  // Phase 2: blocks 25,001-358,333
  if (blockNumber < 358333) return 3e18;

  // Phase 3: blocks 358,334-691,666
  if (blockNumber < 691666) return 1.5e18;

  // Phase 4: blocks 691,667-5,691,666
  if (blockNumber < 5691666) return 1e18;

  // Phase 5: Year 1 remainder
  if (blockNumber < 7884000) return 0.5e18;

  // Year 2
  if (blockNumber < 15768000) return 0.375e18;

  // Year 3
  if (blockNumber < 23652000) return 0.28125e18;

  // Year 4
  if (blockNumber < 31536000) return 0.2109375e18;

  // Year 5
  if (blockNumber < 39420000) return 0.158203125e18;

  // Year 6+: Minimum floor
  return 0.125e18;
}
```

**Supply Projection**:
- **Year 1**: 8.596M HALO (8.6% of max supply)
- **Year 2**: 11.552M HALO cumulative
- **Year 3**: 13.769M HALO cumulative
- **Year 4**: 15.432M HALO cumulative
- **Year 5**: 16.679M HALO cumulative
- **Year 6+**: ~0.986M HALO/year
- **Maximum Supply**: 100,000,000 HALO (reached in ~90 years)

**Key Features**:
- Front-loaded rewards for early network security
- 75% annual reduction after Year 1
- 0.125 HALO/block minimum floor
- 100M hard cap (no infinite inflation)
- Combined with 40% fee burn = deflationary over time

## DAG Parameters (Ethash)

| Parameter | Value | Ethereum |
|-----------|-------|----------|
| **Initial DAG Size** | 512 MB | 1 GB |
| **DAG Growth per Epoch** | 1 MB | 8 MB |
| **Epoch Length** | 30,000 blocks | 30,000 blocks |
| **Initial Cache Size** | 16 MB | 16 MB |
| **Cache Growth per Epoch** | 128 KB | 128 KB |

**DAG Size Over Time** (at 4s blocks):
```
Epoch 0: 512 MB
Epoch 10: 522 MB (~41 days)
Epoch 100: 612 MB (~14 months)
Epoch 1000: 1512 MB (~12 years)
```

## EIP-1559 Parameters

| Parameter | Value | Ethereum |
|-----------|-------|----------|
| **Initial Base Fee** | 1 Gwei | 1 Gwei |
| **BaseFeeChangeDenominator** | 8 | 8 |
| **ElasticityMultiplier** | 2 | 2 |
| **Target Gas per Block** | 75M (50% of limit) | 15M |

### Fee Distribution (BASE FEES ONLY)

| Recipient | Percentage | Purpose |
|-----------|------------|---------|
| **Burned** | 40% | Deflationary mechanism |
| **Miners** | 30% | Mining incentive |
| **Ecosystem Fund** | 20% | Development, grants, marketing |
| **Reserve Fund** | 10% | Emergency reserve, stability |

**Note**: Priority fees (tips) go 100% to miners (not split).

**Formula**:
```javascript
totalBaseFee = baseFeePerGas × gasUsed

burned = totalBaseFee × 40 / 100
minerShare = totalBaseFee × 30 / 100
ecosystemShare = totalBaseFee × 20 / 100
reserveShare = totalBaseFee × 10 / 100

minerTotal = minerShare + priorityFees
```

## Cache and Storage Settings

| Parameter | Value | Ethereum Default |
|-----------|-------|------------------|
| **StateCacheSize** | 1,000,000 entries | Variable |
| **CodeCacheSize** | 100,000 entries | Variable |
| **TrieCleanCacheSize** | 512 MB | 256 MB |
| **TrieDirtyCacheSize** | 256 MB | 256 MB |
| **DatabaseCache** | 2048 MB | 1024 MB |

## Transaction Pool Settings

| Parameter | Value | Ethereum Default |
|-----------|-------|------------------|
| **GlobalSlots** | 8,192 | 4,096 |
| **GlobalQueue** | 4,096 | 1,024 |
| **AccountSlots** | 128 | 16 |
| **AccountQueue** | 64 | 64 |

## Network Settings

| Parameter | Value | Ethereum Default |
|-----------|-------|------------------|
| **MaxMessageSize** | 10 MB | 10 MB |
| **MaxPeers** | 50 | 50 |
| **Network Protocol** | eth/68 | eth/68 |

## EIP Activation (All at Genesis - Block 0)

### Homestead
- ✅ EIP-2 (Homestead gas costs)
- ✅ EIP-7 (DELEGATECALL)

### Tangerine Whistle
- ✅ EIP-150 (Gas cost changes)

### Spurious Dragon
- ✅ EIP-155 (Replay protection)
- ✅ EIP-160 (EXP cost increase)
- ✅ EIP-161 (State trie clearing)
- ✅ EIP-170 (Contract code size limit)

### Byzantium
- ✅ EIP-100 (Difficulty adjustment)
- ✅ EIP-140 (REVERT opcode)
- ✅ EIP-198 (Modular exponentiation)
- ✅ EIP-211 (RETURNDATASIZE)
- ✅ EIP-212 (bn256 precompile)
- ✅ EIP-213 (REVERT)
- ✅ EIP-214 (STATICCALL)
- ✅ EIP-658 (Transaction status)

### Constantinople
- ✅ EIP-145 (Bitwise shifting)
- ✅ EIP-1014 (CREATE2)
- ✅ EIP-1052 (EXTCODEHASH)

### Istanbul
- ✅ EIP-152 (Blake2 precompile)
- ✅ EIP-1108 (Alt_bn128 gas reduction)
- ✅ EIP-1344 (CHAINID opcode)
- ✅ EIP-1884 (Gas costs for state access)
- ✅ EIP-2028 (Calldata gas reduction)
- ✅ EIP-2200 (Net gas metering)

### Berlin
- ✅ EIP-2565 (ModExp gas cost)
- ✅ EIP-2718 (Typed transactions)
- ✅ EIP-2929 (Gas cost increases)
- ✅ EIP-2930 (Access lists)

### London
- ✅ EIP-1559 (Fee market - CUSTOM IMPLEMENTATION)
- ✅ EIP-3198 (BASEFEE opcode)
- ✅ EIP-3529 (Refund reduction)
- ✅ EIP-3541 (Reject contracts starting with 0xEF)

### Shanghai
- ✅ EIP-3651 (Warm COINBASE)
- ✅ EIP-3855 (PUSH0 opcode)
- ✅ EIP-3860 (Limit and meter initcode)

## Command Line Example

### Start Full Node
```bash
geth \
  --datadir ./halo-data \
  --networkid 12000 \
  --port 30303 \
  --http \
  --http.addr 0.0.0.0 \
  --http.port 8545 \
  --http.api eth,net,web3,personal,miner \
  --http.corsdomain "*" \
  --ws \
  --ws.addr 0.0.0.0 \
  --ws.port 8546 \
  --ws.api eth,net,web3 \
  --db.engine pebble \
  --cache 2048 \
  --maxpeers 50 \
  --mine \
  --miner.threads 2 \
  --miner.etherbase YOUR_MINER_ADDRESS \
  --bootnodes "enode://BOOTNODE1,enode://BOOTNODE2"
```

### Start Archive Node
```bash
geth \
  --datadir ./halo-archive \
  --networkid 12000 \
  --gcmode archive \
  --syncmode full \
  --cache 4096 \
  --http \
  --http.api eth,net,web3,debug,trace \
  --db.engine pebble
```

### Start Light Client (Future)
```bash
geth \
  --datadir ./halo-light \
  --networkid 12000 \
  --syncmode light \
  --http
```

## Genesis JSON Template

```json
{
  "config": {
    "chainId": 12000,
    "networkId": 12000,
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
  "nonce": "0x427953706c697473",
  "timestamp": "0x65700000",
  "extraData": "0x48616c6f204e6574776f726b20763120343273",
  "gasLimit": "0x8F0D180",
  "difficulty": "0x10000",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "coinbase": "0x0000000000000000000000000000000000000000",
  "alloc": {},
  "number": "0x0",
  "gasUsed": "0x0",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
```

## RPC Methods (Standard eth_ namespace)

All standard Ethereum RPC methods supported:
- `eth_chainId` → Returns 12000
- `eth_blockNumber` → Current block
- `eth_getBalance` → Account balance
- `eth_sendTransaction` → Send transaction
- `eth_call` → Execute call
- `eth_estimateGas` → Estimate gas
- `eth_getBlockByNumber` → Get block
- `eth_getTransactionReceipt` → Get receipt
- Plus all other standard methods

## Performance Benchmarks (Expected)

| Metric | Target | Notes |
|--------|--------|-------|
| Block Time | 4 seconds ± 500ms | Network dependent |
| Uncle Rate | < 3% | Lower due to longer block time |
| Orphan Rate | < 1% | Monitor closely |
| Sync Speed | > 500 blocks/sec | With fast storage |
| TPS Capacity | ~1000-2000 | With 150M gas, 21k gas/tx |

## Monitoring Key Metrics

1. **Block Production**
   - Average block time: Should be ~4s
   - Uncle rate: Should be < 3%
   - Orphan rate: Should be < 1%

2. **Fee Distribution**
   - Ecosystem fund balance growth
   - Reserve fund balance growth
   - Total supply (after burns)

3. **Network Health**
   - Peer count
   - Sync status
   - Hashrate distribution

4. **Rewards**
   - Current block reward (verify schedule)
   - Uncle rewards paid
   - Nephew rewards paid

## Security Parameters

| Parameter | Value | Purpose |
|-----------|-------|---------|
| **51% Attack Cost** | Variable | Depends on hashrate |
| **Uncle Inclusion** | Max 2 blocks deep | Reduces reorganization risk |
| **Multisig Funds** | Recommended | For ecosystem/reserve |
| **Timelock** | Recommended | For fund withdrawals |

---

**Version**: 1.0.0
**Last Updated**: 2025-01-XX
**Specification**: Complete
**Status**: Ready for Integration Testing
