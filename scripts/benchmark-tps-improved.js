#!/usr/bin/env node
/**
 * Improved Halo Chain TPS Benchmark
 * Better nonce management and higher transaction count
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
const BATCH_SIZE = 5; // Send 5 at a time with proper nonce management

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

    console.log('⚡ Halo Chain TPS Benchmark (Improved)');
    console.log('='.repeat(70));
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
    console.log(`   Batch size: ${BATCH_SIZE} (sequential with nonce tracking)`);
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

    // Get starting nonce
    let currentNonce = await provider.getTransactionCount(wallet.address);
    console.log(`Starting nonce: ${currentNonce}`);
    console.log('');

    // Send transactions with proper nonce management
    console.log('🚀 Sending transactions...');
    const startTime = Date.now();
    let successCount = 0;
    let failCount = 0;
    const txHashes = [];

    for (let i = 0; i < txCount; i += BATCH_SIZE) {
        const batch = [];
        const batchSize = Math.min(BATCH_SIZE, txCount - i);

        for (let j = 0; j < batchSize; j++) {
            const txPromise = wallet.sendTransaction({
                to: wallet.address,
                value: ethers.parseEther('0.001'),
                gasLimit: 21000,
                gasPrice: ethers.parseUnits('2', 'gwei'),
                nonce: currentNonce++
            }).then(tx => {
                txHashes.push(tx.hash);
                successCount++;
                process.stdout.write('.');
                return tx;
            }).catch(error => {
                failCount++;
                process.stdout.write('x');
                console.error(`\n   Error: ${error.message.substring(0, 80)}`);
                return null;
            });

            batch.push(txPromise);
        }

        // Wait for batch to be sent
        await Promise.all(batch);

        // Small delay between batches
        if (i + BATCH_SIZE < txCount) {
            await new Promise(resolve => setTimeout(resolve, 200));
        }

        // Show progress every 20 transactions
        if ((i + BATCH_SIZE) % 20 === 0) {
            console.log(` [${i + BATCH_SIZE}/${txCount}]`);
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
    console.log('⏳ Waiting for all confirmations...');
    process.stdout.write('   ');

    let confirmed = 0;
    const confirmStart = Date.now();
    const timeout = 120000; // 2 minutes

    for (const hash of txHashes) {
        if (hash) {
            try {
                await provider.waitForTransaction(hash, 1, timeout);
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
    const totalTime = (Date.now() - startTime) / 1000;
    console.log('');
    console.log(`✅ Confirmed: ${confirmed}/${successCount}`);
    console.log(`   Confirmation time: ${confirmTime.toFixed(2)}s`);
    console.log(`   Total time: ${totalTime.toFixed(2)}s`);
    console.log('');

    // Get final stats
    const blockAfter = await provider.getBlockNumber();
    const blocksUsed = blockAfter - blockBefore;

    console.log('📊 Final Statistics:');
    console.log(`   Starting block: ${blockBefore}`);
    console.log(`   Final block: ${blockAfter}`);
    console.log(`   Blocks used: ${blocksUsed}`);
    console.log(`   Avg tx per block: ${(confirmed / blocksUsed).toFixed(2)}`);
    console.log(`   Block time: ~1 second (target)`);
    console.log(`   Achieved TPS: ${(confirmed / blocksUsed).toFixed(2)}`);
    console.log(`   Overall throughput: ${(confirmed / totalTime).toFixed(2)} tx/s`);
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
    let feeDistributionCorrect = false;
    if (ecosystemDiff > 0n && reserveDiff > 0n) {
        const ratio = Number(ecosystemDiff * 1000n / reserveDiff) / 1000;
        console.log('📊 Fee Distribution Verification:');
        console.log(`   Ecosystem/Reserve ratio: ${ratio.toFixed(3)}`);
        console.log(`   Expected: 2.000 (20%/10%)`);
        console.log(`   Difference: ${Math.abs(ratio - 2.0).toFixed(3)}`);

        if (ratio > 1.8 && ratio < 2.2) {
            console.log('   ✅ Fee distribution is CORRECT!');
            feeDistributionCorrect = true;
        } else {
            console.log('   ⚠️  Fee distribution ratio is off');
        }
    } else {
        console.log('⚠️  Insufficient fees distributed for ratio calculation');
    }

    // Save results to JSON
    const results = {
        test_name: "Halo Chain TPS Benchmark (Improved)",
        test_date: new Date().toISOString(),
        configuration: {
            total_transactions_requested: txCount,
            batch_size: BATCH_SIZE,
            transaction_type: "self_transfer",
            value_per_tx: "0.001 HALO",
            gas_limit: 21000,
            gas_price: "2 gwei"
        },
        results: {
            transactions_sent: successCount,
            transactions_failed: failCount,
            transactions_confirmed: confirmed,
            success_rate_percent: ((successCount / txCount) * 100).toFixed(2),
            confirmation_rate_percent: confirmed > 0 ? ((confirmed / successCount) * 100).toFixed(2) : 0,
            send_time_seconds: sendTime.toFixed(2),
            confirmation_time_seconds: confirmTime.toFixed(2),
            total_time_seconds: totalTime.toFixed(2),
            send_rate_tps: (successCount / sendTime).toFixed(2),
            blocks_used: blocksUsed,
            avg_tx_per_block: (confirmed / blocksUsed).toFixed(2),
            achieved_tps: (confirmed / blocksUsed).toFixed(2),
            overall_throughput_tps: (confirmed / totalTime).toFixed(2)
        },
        fee_distribution: {
            ecosystem_fund: {
                address: ECOSYSTEM_ADDR,
                balance_before: ethers.formatEther(ecosystemBefore),
                balance_after: ethers.formatEther(ecosystemAfter),
                balance_change: ethers.formatEther(ecosystemDiff),
                balance_change_wei: ecosystemDiff.toString()
            },
            reserve_fund: {
                address: RESERVE_ADDR,
                balance_before: ethers.formatEther(reserveBefore),
                balance_after: ethers.formatEther(reserveAfter),
                balance_change: ethers.formatEther(reserveDiff),
                balance_change_wei: reserveDiff.toString()
            },
            miner: {
                address: MINER_ADDR,
                balance_before: ethers.formatEther(minerBefore),
                balance_after: ethers.formatEther(minerAfter),
                balance_change: ethers.formatEther(minerDiff)
            },
            verification: {
                ecosystem_reserve_ratio: ecosystemDiff > 0n && reserveDiff > 0n ?
                    (Number(ecosystemDiff * 1000n / reserveDiff) / 1000) : 0,
                expected_ratio: 2.0,
                ratio_correct: feeDistributionCorrect,
                status: feeDistributionCorrect ? "PASSED" : "NEEDS_REVIEW"
            }
        },
        network_stats: {
            chain_id: 12000,
            starting_block: blockBefore,
            ending_block: blockAfter,
            blocks_produced: blocksUsed
        }
    };

    const outputFile = 'TPS_BENCHMARK_RESULTS.json';
    fs.writeFileSync(outputFile, JSON.stringify(results, null, 2));

    console.log('');
    console.log('='.repeat(70));
    console.log('✅ Benchmark complete!');
    console.log(`   Results saved to: ${outputFile}`);
    console.log('');

    // Summary
    console.log('📋 SUMMARY:');
    console.log(`   ✅ Sent: ${successCount}/${txCount} transactions`);
    console.log(`   ✅ Confirmed: ${confirmed} transactions`);
    console.log(`   ✅ TPS: ${(confirmed / blocksUsed).toFixed(2)}`);
    console.log(`   ${feeDistributionCorrect ? '✅' : '⚠️ '} Fee distribution: ${feeDistributionCorrect ? 'CORRECT' : 'NEEDS REVIEW'}`);
}

main().catch(error => {
    console.error('❌ Error:', error.message);
    process.exit(1);
});
