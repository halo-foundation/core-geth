#!/usr/bin/env node
/**
 * Halo Chain TPS Benchmark
 * Measures transactions per second and tests fee distribution
 */

const { ethers } = require('ethers');
const fs = require('fs');
const path = require('path');

// Configuration
const RPC_URL = 'http://localhost:8545';
const ECOSYSTEM_ADDR = '0xa7548DF196e2C1476BDc41602E288c0A8F478c4f';
const RESERVE_ADDR = '0xb95ae9b737e104C666d369CFb16d6De88208Bd80';
const MINER_ADDR = '0x69AEd36e548525ED741052A6572Bb1328973b44F';

// Benchmark settings
const DEFAULT_TX_COUNT = 100;
const DEFAULT_CONCURRENT = 10;

async function loadMinerWallet(provider) {
    const keystorePath = path.join(__dirname, '..', 'halo-test', 'keystore');
    const files = fs.readdirSync(keystorePath);
    const keystoreFile = files.find(f => f.startsWith('UTC--'));

    if (!keystoreFile) {
        throw new Error('Miner keystore not found');
    }

    const json = fs.readFileSync(path.join(keystorePath, keystoreFile), 'utf8');
    const wallet = await ethers.Wallet.fromEncryptedJson(json, 'test123');
    return wallet.connect(provider);
}

async function main() {
    const args = process.argv.slice(2);
    const txCount = parseInt(args[0]) || DEFAULT_TX_COUNT;
    const concurrent = parseInt(args[1]) || DEFAULT_CONCURRENT;

    console.log('⚡ Halo Chain TPS Benchmark');
    console.log('='.repeat(60));
    console.log('');

    // Connect
    const provider = new ethers.JsonRpcProvider(RPC_URL);

    try {
        const network = await provider.getNetwork();
        console.log('📡 Connected to Halo Chain');
        console.log(`   Chain ID: ${network.chainId}`);
        console.log('');
    } catch (error) {
        console.error('❌ Cannot connect to node');
        process.exit(1);
    }

    // Load wallet
    console.log('🔐 Loading miner wallet...');
    const wallet = await loadMinerWallet(provider);
    console.log(`   Address: ${wallet.address}`);

    const balance = await provider.getBalance(wallet.address);
    console.log(`   Balance: ${ethers.formatEther(balance)} HALO`);
    console.log('');

    // Benchmark configuration
    console.log('⚙️  Benchmark Configuration:');
    console.log(`   Total transactions: ${txCount}`);
    console.log(`   Concurrent sends: ${concurrent}`);
    console.log(`   Target: Self-transfers (minimal gas)`);
    console.log('');

    // Get initial balances
    console.log('📊 Initial Fund Balances:');
    const ecosystemBefore = await provider.getBalance(ECOSYSTEM_ADDR);
    const reserveBefore = await provider.getBalance(RESERVE_ADDR);
    const minerBefore = await provider.getBalance(MINER_ADDR);

    console.log(`   Ecosystem: ${ethers.formatEther(ecosystemBefore)} HALO`);
    console.log(`   Reserve:   ${ethers.formatEther(reserveBefore)} HALO`);
    console.log(`   Miner:     ${ethers.formatEther(minerBefore)} HALO`);
    console.log('');

    const blockBefore = await provider.getBlockNumber();
    console.log(`📦 Starting Block: ${blockBefore}`);
    console.log('');

    // Send transactions
    console.log('🚀 Sending transactions...');
    const startTime = Date.now();
    let successCount = 0;
    let failCount = 0;
    const txHashes = [];

    // Send in batches
    for (let i = 0; i < txCount; i += concurrent) {
        const batch = [];
        const batchSize = Math.min(concurrent, txCount - i);

        for (let j = 0; j < batchSize; j++) {
            const txPromise = wallet.sendTransaction({
                to: wallet.address, // Send to self
                value: ethers.parseEther('0.001'),
                gasLimit: 21000,
                gasPrice: ethers.parseUnits('2', 'gwei')
            }).then(tx => {
                txHashes.push(tx.hash);
                successCount++;
                process.stdout.write('.');
                return tx;
            }).catch(error => {
                failCount++;
                process.stdout.write('x');
                return null;
            });

            batch.push(txPromise);
        }

        await Promise.all(batch);

        // Small delay between batches
        if (i + concurrent < txCount) {
            await new Promise(resolve => setTimeout(resolve, 100));
        }
    }

    console.log('');
    const sendTime = (Date.now() - startTime) / 1000;
    console.log('');
    console.log(`✅ Transactions sent: ${successCount}/${txCount}`);
    console.log(`   Failed: ${failCount}`);
    console.log(`   Time: ${sendTime.toFixed(2)}s`);
    console.log(`   Send rate: ${(successCount / sendTime).toFixed(2)} tx/s`);
    console.log('');

    // Wait for confirmations
    console.log('⏳ Waiting for confirmations...');
    process.stdout.write('   ');

    let confirmed = 0;
    const confirmStart = Date.now();

    for (const hash of txHashes) {
        if (hash) {
            try {
                await provider.waitForTransaction(hash, 1, 60000);
                confirmed++;
                if (confirmed % 10 === 0) {
                    process.stdout.write(`${confirmed} `);
                }
            } catch (error) {
                // Timeout or error
            }
        }
    }

    console.log('');
    const confirmTime = (Date.now() - confirmStart) / 1000;
    console.log('');
    console.log(`✅ Confirmed: ${confirmed}/${successCount}`);
    console.log(`   Time: ${confirmTime.toFixed(2)}s`);
    console.log('');

    // Get final stats
    const blockAfter = await provider.getBlockNumber();
    const blocksUsed = blockAfter - blockBefore;

    console.log('📊 Final Statistics:');
    console.log(`   Final block: ${blockAfter}`);
    console.log(`   Blocks used: ${blocksUsed}`);
    console.log(`   Avg tx per block: ${(confirmed / blocksUsed).toFixed(2)}`);
    console.log(`   Block time: ~1 second`);
    console.log(`   Achieved TPS: ${(confirmed / blocksUsed).toFixed(2)}`);
    console.log(`   Theoretical max: ${confirmed / blocksUsed} tx/s`);
    console.log('');

    // Check fee distribution
    console.log('💰 Fee Distribution Results:');
    const ecosystemAfter = await provider.getBalance(ECOSYSTEM_ADDR);
    const reserveAfter = await provider.getBalance(RESERVE_ADDR);
    const minerAfter = await provider.getBalance(MINER_ADDR);

    const ecosystemDiff = ecosystemAfter - ecosystemBefore;
    const reserveDiff = reserveAfter - reserveBefore;
    const minerDiff = minerAfter - minerBefore;

    console.log(`   Ecosystem: ${ethers.formatEther(ecosystemAfter)} HALO (+${ethers.formatEther(ecosystemDiff)})`);
    console.log(`   Reserve:   ${ethers.formatEther(reserveAfter)} HALO (+${ethers.formatEther(reserveDiff)})`);
    console.log(`   Miner:     ${ethers.formatEther(minerAfter)} HALO (+${ethers.formatEther(minerDiff)})`);
    console.log('');

    // Verify fee distribution
    if (ecosystemDiff > 0n && reserveDiff > 0n) {
        const ratio = Number(ecosystemDiff * 1000n / reserveDiff) / 1000;
        console.log('📊 Fee Distribution Verification:');
        console.log(`   Ecosystem/Reserve ratio: ${ratio.toFixed(3)}`);
        console.log(`   Expected: 2.000 (20%/10%)`);

        if (ratio > 1.8 && ratio < 2.2) {
            console.log('   ✅ Fee distribution is correct!');
        } else {
            console.log('   ⚠️  Fee distribution ratio is off');
        }
    } else {
        console.log('⚠️  No fees distributed (may need more transactions)');
    }

    console.log('');
    console.log('='.repeat(60));
    console.log('✅ Benchmark complete!');
}

main().catch(error => {
    console.error('❌ Error:', error.message);
    process.exit(1);
});
