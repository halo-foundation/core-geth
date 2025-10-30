# Halo Chain - Integration TODOs

⚠️ **WARNING**: This document contains outdated reward schedule information. Current active parameters: 40/3/1.5/1/0.5 HALO rewards, 8.596M Year 1 supply, 100M max. See HALO_PARAMETERS.md for accurate data.


This document outlines the remaining integration work needed to fully activate Halo chain features.

## Critical Integrations Required

### 1. MaxUncles Chain-Aware Configuration

**File**: `consensus/ethash/consensus.go`
**Line**: ~47
**Current Code**:
```go
var (
	maxUncles              = 2                // Maximum number of uncles allowed in a single block
	allowedFutureBlockTime = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
)
```

**Required Change**:
```go
var (
	allowedFutureBlockTime = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
)

// getMaxUncles returns the maximum number of uncles allowed based on chain configuration
func getMaxUncles(chainID *big.Int) int {
	if chainID != nil && chainID.Uint64() == 12000 {
		return 1 // Halo allows max 1 uncle
	}
	return 2 // Default Ethereum allows 2 uncles
}
```

**Then Update**: All references to `maxUncles` to use `getMaxUncles(chainID)` instead

**Locations to Update**:
1. `verifyHeader()` function - uncle count validation
2. `VerifyUncles()` function - uncle verification
3. Any other uncle validation logic

### 2. EIP-1559 Fee Distribution Integration

**File**: `consensus/ethash/consensus.go`
**Function**: `Finalize()`
**Location**: After transaction processing, before state root calculation

**Add Import**:
```go
import (
	// ... existing imports
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
)
```

**Add Distribution Call**:
```go
func (ethash *Ethash) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, withdrawals []*types.Withdrawal) {
	// ... existing code ...

	// Accumulate block rewards
	mutations.AccumulateRewards(chain.Config(), state, header, uncles)

	// Halo: Apply custom EIP-1559 fee distribution
	if chain.Config().GetChainID() != nil && chain.Config().GetChainID().Uint64() == 12000 {
		if header.BaseFee != nil {
			err := eip1559.ApplyHaloBaseFeeDistribution(state, header, header.BaseFee, header.GasUsed)
			if err != nil {
				// Log error but don't fail block (addresses should be validated at genesis)
				log.Error("Failed to apply Halo fee distribution", "err", err, "block", header.Number)
			}
		}
	}

	// ... rest of finalization ...
}
```

**Important**: This needs to be called AFTER `AccumulateRewards` but BEFORE the final state root is calculated.

### 3. Uncle Depth Validation

**File**: `consensus/ethash/consensus.go`
**Function**: `VerifyUncles()`

**Add Depth Check** for Halo:
```go
func (ethash *Ethash) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	// ... existing code ...

	// Halo-specific: Validate uncle depth (max 2 blocks)
	if chain.Config().GetChainID() != nil && chain.Config().GetChainID().Uint64() == 12000 {
		for _, uncle := range block.Uncles() {
			depth := new(big.Int).Sub(block.Number(), uncle.Number)
			if depth.Uint64() > 2 {
				return fmt.Errorf("uncle too deep: depth %d, max allowed 2", depth.Uint64())
			}
		}
	}

	// ... rest of validation ...
}
```

### 4. Custom DAG Parameters (Optional - If Needed)

**File**: `consensus/ethash/algorithm.go`

**Current**:
```go
const (
	datasetInitBytes   = 1 << 30 // 1GB
	datasetGrowthBytes = 1 << 23 // 8MB
)
```

**For Halo-specific DAG** (if required):
```go
// getDatasetParams returns DAG parameters based on chain
func getDatasetParams(chainID *big.Int) (uint64, uint64) {
	if chainID != nil && chainID.Uint64() == 12000 {
		return 512 * 1024 * 1024, 1 * 1024 * 1024 // 512MB initial, 1MB growth
	}
	return 1 << 30, 1 << 23 // Default: 1GB initial, 8MB growth
}

// calcDatasetSize calculates the dataset size for epoch
func calcDatasetSize(epoch uint64, chainID *big.Int) uint64 {
	initBytes, growthBytes := getDatasetParams(chainID)
	size := initBytes + growthBytes*epoch - mixBytes
	for !new(big.Int).SetUint64(size / mixBytes).ProbablyPrime(1) {
		size -= 2 * mixBytes
	}
	return size
}
```

**Note**: This change requires passing chainID through multiple layers. May not be necessary if standard DAG is acceptable.

## Configuration Changes

### 5. Set Fund Addresses

**File**: `params/genesis_halo.go`
**Lines**: 18-24

**Current (INSECURE)**:
```go
var (
	HaloEcosystemFundAddress = common.HexToAddress("0x0000000000000000000000000000000000000001")
	HaloReserveFundAddress = common.HexToAddress("0x0000000000000000000000000000000000000002")
)
```

**Replace with ACTUAL addresses**:
```go
var (
	// Ecosystem Fund - Receives 20% of base fees
	// CRITICAL: Set to actual multisig address before deployment
	HaloEcosystemFundAddress = common.HexToAddress("0xYOUR_ECOSYSTEM_MULTISIG_ADDRESS")

	// Reserve Fund - Receives 10% of base fees
	// CRITICAL: Set to actual multisig address before deployment
	HaloReserveFundAddress = common.HexToAddress("0xYOUR_RESERVE_MULTISIG_ADDRESS")
)
```

**Validation**: The `ApplyHaloBaseFeeDistribution()` function will error if these are zero addresses.

### 6. Set Genesis Timestamp

**File**: `params/genesis_halo.go`
**Line**: ~46

**Current**:
```go
Timestamp: 1700000000, // TODO: Set to actual launch timestamp
```

**Set to Launch Time**:
```go
Timestamp: 1740787200, // Example: March 1, 2025 00:00:00 UTC
```

### 7. Update Genesis Hash

**After first initialization**, get the genesis hash and update:

**File**: `params/genesis_halo.go`
**Line**: ~16

```bash
# Initialize genesis
./build/bin/geth --datadir=./halo-data init halo_genesis.json

# Get hash
./build/bin/geth --datadir=./halo-data --exec 'eth.getBlock(0).hash' console
```

**Update**:
```go
var HaloGenesisHash = common.HexToHash("0xACTUAL_GENESIS_HASH_FROM_INIT")
```

### 8. Configure Bootnodes

**File**: `params/bootnodes_halo.go`

**After setting up nodes**, get enode URLs:
```bash
./build/bin/geth --datadir=./halo-node1 --exec 'admin.nodeInfo.enode' attach
```

**Update**:
```go
var HaloBootnodes = []string{
	"enode://abc123...@203.0.113.1:30303",
	"enode://def456...@203.0.113.2:30303",
	"enode://ghi789...@203.0.113.3:30303",
}
```

## Testing Requirements

### Unit Tests (Already Created ✅)
```bash
go test ./params -v -run Halo
go test ./params/mutations -v -run Halo
```

### Integration Tests (TODO)

#### Test 1: Block Reward Distribution
```bash
# Start node and mine blocks
./build/bin/geth --datadir=./test-data --networkid 12000 --mine --miner.threads=1

# Verify rewards at milestones
# Block 0: 5 HALO
# Block 100,000: 4 HALO
# Block 400,000: 3 HALO
# Block 700,000: 2 HALO
```

#### Test 2: EIP-1559 Fee Distribution
```javascript
// In geth console
// Send transaction with base fee
eth.sendTransaction({from: eth.accounts[0], to: "0x...", value: web3.toWei(1, "ether")})

// Check fund balances increased
eth.getBalance("ECOSYSTEM_FUND_ADDRESS") // Should show 20% of base fees
eth.getBalance("RESERVE_FUND_ADDRESS")   // Should show 10% of base fees
```

#### Test 3: Uncle Rewards
```javascript
// Mine blocks with uncles
miner.start(2) // Start multiple threads to create uncles

// Verify uncle was included
eth.getBlock(NUMBER).uncles.length // Should be 0 or 1

// Check uncle received 50% or 37.5% of block reward
```

#### Test 4: Network Sync
```bash
# Start bootnode
./build/bin/geth --datadir=./node1 --networkid 12000 --mine

# Start peer node
./build/bin/geth --datadir=./node2 --networkid 12000 --bootnodes="enode://..."

# Verify sync
# Both nodes should show same block height
```

## Integration Checklist

- [ ] Implement `getMaxUncles()` function
- [ ] Update all uncle count validations to use `getMaxUncles()`
- [ ] Add EIP-1559 distribution call to `Finalize()`
- [ ] Add uncle depth validation for Halo
- [ ] Set ecosystem fund address (multisig)
- [ ] Set reserve fund address (multisig)
- [ ] Set genesis timestamp to launch time
- [ ] Initialize genesis and get hash
- [ ] Update HaloGenesisHash
- [ ] Set up bootnodes
- [ ] Update HaloBootnodes array
- [ ] Run all unit tests
- [ ] Test block reward schedule
- [ ] Test EIP-1559 fee distribution
- [ ] Test uncle inclusion and rewards
- [ ] Test multi-node network
- [ ] Deploy testnet
- [ ] Security audit of fund addresses
- [ ] Deploy mainnet

## Priority Order

### Phase 1: Core Integration (CRITICAL)
1. Implement maxUncles chain detection
2. Integrate EIP-1559 fee distribution
3. Add uncle depth validation
4. Set actual fund addresses

### Phase 2: Configuration (HIGH)
5. Set genesis timestamp
6. Initialize genesis block
7. Update genesis hash
8. Set up initial bootnodes

### Phase 3: Testing (HIGH)
9. Run all unit tests
10. Integration test: reward schedule
11. Integration test: fee distribution
12. Integration test: network sync

### Phase 4: Deployment (MEDIUM)
13. Deploy testnet
14. Monitor testnet for issues
15. Security audit
16. Deploy mainnet

## Code Review Points

When implementing integrations, ensure:

1. **Thread Safety**: EIP-1559 distribution happens in Finalize(), check for race conditions
2. **Error Handling**: Distribution errors should log but not fail blocks (after addresses validated)
3. **Gas Accounting**: Fee distribution doesn't consume gas from block limit
4. **State Consistency**: Distribution happens before state root calculation
5. **Backward Compatibility**: Changes don't affect other chains (use chainID checks)

## Testing Scenarios

### Edge Cases to Test

1. **Block with no transactions**: EIP-1559 distribution should handle 0 fees
2. **Block with max uncles**: Should only allow 1 uncle for Halo
3. **Uncle at max depth**: Depth 2 should be accepted, depth 3 rejected
4. **Zero fund addresses**: Should error (caught by validation)
5. **Overflow/Underflow**: Fee distribution math should handle large numbers
6. **Genesis block**: Reward should be 5 HALO
7. **Block 100,000**: Reward should transition to 4 HALO
8. **Network partition**: Nodes should re-sync correctly

## Documentation to Update

After integration:

1. Update README with Halo chain support
2. Add Halo to supported networks list
3. Create MetaMask/wallet integration guide
4. Document RPC endpoints
5. Create block explorer integration guide

## Support Infrastructure Needed

1. **Block Explorer**: Modified for Halo's fee distribution
2. **Faucet**: For testnet
3. **Monitoring**: Track fund balances, uncle rate, orphan rate
4. **Documentation Site**: Setup guides, API docs
5. **Discord/Forum**: Community support

---

**Next Steps**:
1. Review this document with team
2. Assign integration tasks
3. Set timeline for each phase
4. Begin Phase 1 implementation

**Estimated Time**:
- Phase 1: 2-3 days
- Phase 2: 1 day
- Phase 3: 3-5 days
- Phase 4: Ongoing

**Blockers**: None identified
**Dependencies**: Need actual multisig addresses for funds before testnet deployment
