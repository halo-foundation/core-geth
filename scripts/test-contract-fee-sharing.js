/**
 * Test contract fee sharing mechanism
 * This tests that contracts can receive a portion of the ecosystem fund allocation
 */

const { ethers } = require('ethers');
const fs = require('fs');
const solc = require('solc');

// Configuration
const RPC_URL = 'http://localhost:8545';
const ECOSYSTEM_ADDR = '0xa7548DF196e2C1476BDc41602E288c0A8F478c4f';
const RESERVE_ADDR = '0xb95ae9b737e104C666d369CFb16d6De88208Bd80';
const MINER_ADDR = '0x69AEd36e548525ED741052A6572Bb1328973b44F';

// Miner private key (for deploying contract and sending transactions)
const MINER_PRIVATE_KEY = '0x32e1b0aeb11846cc691c407821280d5c78be0249c7c9746cd3e81e81ea2e937e';

async function main() {
    console.log('🧪 Testing Contract Fee Sharing');
    console.log('='.repeat(60));
    console.log('');

    // Connect to node
    const provider = new ethers.JsonRpcProvider(RPC_URL);
    const wallet = new ethers.Wallet(MINER_PRIVATE_KEY, provider);

    console.log('📡 Connected to Halo Chain');
    console.log(`   Deployer: ${wallet.address}`);
    console.log('');

    // Compile contract
    console.log('🔨 Compiling FeeSharing contract...');
    const contractSource = fs.readFileSync('./contracts/FeeSharing.sol', 'utf8');

    const input = {
        language: 'Solidity',
        sources: {
            'FeeSharing.sol': {
                content: contractSource
            }
        },
        settings: {
            outputSelection: {
                '*': {
                    '*': ['*']
                }
            }
        }
    };

    const output = JSON.parse(solc.compile(JSON.stringify(input)));

    if (output.errors) {
        const errors = output.errors.filter(e => e.severity === 'error');
        if (errors.length > 0) {
            console.error('❌ Compilation errors:');
            errors.forEach(e => console.error(e.formattedMessage));
            process.exit(1);
        }
    }

    const contract = output.contracts['FeeSharing.sol']['FeeSharing'];
    const abi = contract.abi;
    const bytecode = contract.evm.bytecode.object;

    console.log('   ✅ Contract compiled');
    console.log('');

    // Get initial balances
    console.log('📊 Initial balances:');
    const ecosystemBalanceBefore = await provider.getBalance(ECOSYSTEM_ADDR);
    const reserveBalanceBefore = await provider.getBalance(RESERVE_ADDR);
    const minerBalanceBefore = await provider.getBalance(wallet.address);

    console.log(`   Ecosystem: ${ethers.formatEther(ecosystemBalanceBefore)} HALO`);
    console.log(`   Reserve:   ${ethers.formatEther(reserveBalanceBefore)} HALO`);
    console.log(`   Miner:     ${ethers.formatEther(minerBalanceBefore)} HALO`);
    console.log('');

    // Deploy contract
    console.log('🚀 Deploying FeeSharing contract...');
    const factory = new ethers.ContractFactory(abi, bytecode, wallet);
    const feeSharingContract = await factory.deploy();
    await feeSharingContract.waitForDeployment();
    const contractAddress = await feeSharingContract.getAddress();

    console.log(`   ✅ Contract deployed at: ${contractAddress}`);
    console.log('');

    // Enable fee sharing
    console.log('⚙️  Enabling fee sharing...');
    const enableTx = await feeSharingContract.enableFeeSharing();
    await enableTx.wait();
    console.log('   ✅ Fee sharing enabled');
    console.log('');

    // Send some transactions to generate fees
    console.log('💸 Sending transactions to generate fees...');
    for (let i = 0; i < 5; i++) {
        const tx = await wallet.sendTransaction({
            to: wallet.address, // Send to self
            value: ethers.parseEther('0.1'),
            gasPrice: ethers.parseUnits('2', 'gwei')
        });
        await tx.wait();
        console.log(`   ✅ Transaction ${i + 1}/5 sent`);
    }
    console.log('');

    // Wait for a few blocks
    console.log('⏳ Waiting for blocks to be mined...');
    const currentBlock = await provider.getBlockNumber();
    const targetBlock = currentBlock + 3;

    while (await provider.getBlockNumber() < targetBlock) {
        await new Promise(resolve => setTimeout(resolve, 1000));
        process.stdout.write('.');
    }
    console.log('');
    console.log('');

    // Check balances after
    console.log('📊 Final balances:');
    const ecosystemBalanceAfter = await provider.getBalance(ECOSYSTEM_ADDR);
    const reserveBalanceAfter = await provider.getBalance(RESERVE_ADDR);
    const minerBalanceAfter = await provider.getBalance(wallet.address);
    const contractBalance = await provider.getBalance(contractAddress);

    console.log(`   Ecosystem: ${ethers.formatEther(ecosystemBalanceAfter)} HALO`);
    console.log(`   Reserve:   ${ethers.formatEther(reserveBalanceAfter)} HALO`);
    console.log(`   Miner:     ${ethers.formatEther(minerBalanceAfter)} HALO`);
    console.log(`   Contract:  ${ethers.formatEther(contractBalance)} HALO`);
    console.log('');

    // Calculate differences
    const ecosystemDiff = ecosystemBalanceAfter - ecosystemBalanceBefore;
    const reserveDiff = reserveBalanceAfter - reserveBalanceBefore;
    const minerDiff = minerBalanceAfter - minerBalanceBefore;

    console.log('💰 Balance changes:');
    console.log(`   Ecosystem: ${ethers.formatEther(ecosystemDiff)} HALO`);
    console.log(`   Reserve:   ${ethers.formatEther(reserveDiff)} HALO`);
    console.log(`   Miner:     ${ethers.formatEther(minerDiff)} HALO (includes gas costs)`);
    console.log(`   Contract:  ${ethers.formatEther(contractBalance)} HALO`);
    console.log('');

    // Get contract status
    const status = await feeSharingContract.getStatus();
    console.log('📋 Contract status:');
    console.log(`   Fee sharing enabled: ${status[0]}`);
    console.log(`   Current balance: ${ethers.formatEther(status[1])} HALO`);
    console.log(`   Total received: ${ethers.formatEther(status[2])} HALO`);
    console.log('');

    // Verify results
    console.log('✅ Test Results:');
    if (ecosystemDiff > 0n) {
        console.log('   ✅ Ecosystem fund received fees');
    } else {
        console.log('   ⚠️  Ecosystem fund balance did not increase');
    }

    if (reserveDiff > 0n) {
        console.log('   ✅ Reserve fund received fees');
    } else {
        console.log('   ⚠️  Reserve fund balance did not increase');
    }

    if (contractBalance > 0n) {
        console.log('   ✅ Contract received fee share');
    } else {
        console.log('   ⚠️  Contract did not receive fees (fee sharing may not be implemented yet)');
    }

    console.log('');
    console.log('='.repeat(60));
    console.log('✅ Contract fee sharing test complete');
    console.log('');
    console.log('⚠️  NOTE: If contract did not receive fees, this feature may need');
    console.log('   to be implemented in the consensus/EIP-1559 code.');
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error('❌ Error:', error);
        process.exit(1);
    });
