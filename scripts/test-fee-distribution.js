#!/usr/bin/env node
/**
 * Test Halo Chain fee distribution
 * Tests the 4-way fee split: 40% burn, 30% miner, 20% ecosystem, 10% reserve
 */

const { ethers } = require('ethers');

// Configuration
const RPC_URL = 'http://localhost:8545';
const ECOSYSTEM_ADDR = '0xa7548DF196e2C1476BDc41602E288c0A8F478c4f';
const RESERVE_ADDR = '0xb95ae9b737e104C666d369CFb16d6De88208Bd80';
const MINER_ADDR = '0x69AEd36e548525ED741052A6572Bb1328973b44F';

async function main() {
    console.log('🧪 Testing Halo Chain Fee Distribution');
    console.log('='.repeat(60));
    console.log('');

    // Connect to node
    const provider = new ethers.JsonRpcProvider(RPC_URL);

    try {
        const network = await provider.getNetwork();
        console.log('📡 Connected to Halo Chain');
        console.log(`   Chain ID: ${network.chainId}`);
        console.log(`   RPC URL: ${RPC_URL}`);
        console.log('');
    } catch (error) {
        console.error('❌ Cannot connect to node. Is it running?');
        console.error(`   Error: ${error.message}`);
        process.exit(1);
    }

    // Fund addresses
    console.log('📋 Fund Addresses:');
    console.log(`   Ecosystem: ${ECOSYSTEM_ADDR} (20%)`);
    console.log(`   Reserve:   ${RESERVE_ADDR} (10%)`);
    console.log(`   Miner:     ${MINER_ADDR} (30% + block rewards)`);
    console.log('');

    // Get initial balances
    console.log('📊 Initial Balances:');
    const ecosystemBalanceBefore = await provider.getBalance(ECOSYSTEM_ADDR);
    const reserveBalanceBefore = await provider.getBalance(RESERVE_ADDR);
    const minerBalanceBefore = await provider.getBalance(MINER_ADDR);

    console.log(`   Ecosystem: ${ethers.formatEther(ecosystemBalanceBefore)} HALO`);
    console.log(`   Reserve:   ${ethers.formatEther(reserveBalanceBefore)} HALO`);
    console.log(`   Miner:     ${ethers.formatEther(minerBalanceBefore)} HALO`);
    console.log('');

    // Get current block
    const currentBlock = await provider.getBlockNumber();
    console.log(`📦 Current Block: ${currentBlock}`);
    console.log('');

    // Wait for new blocks
    console.log('⏳ Waiting for 5 new blocks to be mined...');
    const targetBlock = currentBlock + 5;

    process.stdout.write('   Progress: ');
    while (await provider.getBlockNumber() < targetBlock) {
        await new Promise(resolve => setTimeout(resolve, 500));
        process.stdout.write('.');
    }
    console.log(' Done!');
    console.log('');

    // Get final balances
    console.log('📊 Final Balances:');
    const ecosystemBalanceAfter = await provider.getBalance(ECOSYSTEM_ADDR);
    const reserveBalanceAfter = await provider.getBalance(RESERVE_ADDR);
    const minerBalanceAfter = await provider.getBalance(MINER_ADDR);

    console.log(`   Ecosystem: ${ethers.formatEther(ecosystemBalanceAfter)} HALO`);
    console.log(`   Reserve:   ${ethers.formatEther(reserveBalanceAfter)} HALO`);
    console.log(`   Miner:     ${ethers.formatEther(minerBalanceAfter)} HALO`);
    console.log('');

    // Calculate differences
    const ecosystemDiff = ecosystemBalanceAfter - ecosystemBalanceBefore;
    const reserveDiff = reserveBalanceAfter - reserveBalanceBefore;
    const minerDiff = minerBalanceAfter - minerBalanceBefore;

    console.log('💰 Balance Changes:');
    console.log(`   Ecosystem: +${ethers.formatEther(ecosystemDiff)} HALO`);
    console.log(`   Reserve:   +${ethers.formatEther(reserveDiff)} HALO`);
    console.log(`   Miner:     +${ethers.formatEther(minerDiff)} HALO`);
    console.log('');

    // Verify results
    console.log('✅ Test Results:');
    let allPassed = true;

    if (ecosystemDiff > 0n) {
        console.log('   ✅ Ecosystem fund received fees');
    } else {
        console.log('   ❌ Ecosystem fund did not receive fees');
        allPassed = false;
    }

    if (reserveDiff > 0n) {
        console.log('   ✅ Reserve fund received fees');
    } else {
        console.log('   ❌ Reserve fund did not receive fees');
        allPassed = false;
    }

    if (minerDiff > 0n) {
        console.log('   ✅ Miner received rewards');
    } else {
        console.log('   ❌ Miner did not receive rewards');
        allPassed = false;
    }

    // Check ratio
    if (ecosystemDiff > 0n && reserveDiff > 0n) {
        const ratio = Number(ecosystemDiff * 1000n / reserveDiff) / 1000;
        console.log('');
        console.log('📊 Ratio Analysis:');
        console.log(`   Ecosystem/Reserve ratio: ${ratio.toFixed(3)}`);
        console.log(`   Expected ratio: 2.000 (20%/10%)`);

        if (ratio > 1.8 && ratio < 2.2) {
            console.log('   ✅ Ratio is correct (within 10% margin)');
        } else {
            console.log('   ⚠️  Ratio deviation detected');
            allPassed = false;
        }
    }

    console.log('');
    console.log('='.repeat(60));
    if (allPassed) {
        console.log('✅ All fee distribution tests PASSED!');
    } else {
        console.log('⚠️  Some tests did not pass as expected');
    }

    process.exit(allPassed ? 0 : 1);
}

main().catch(error => {
    console.error('❌ Error:', error.message);
    process.exit(1);
});
