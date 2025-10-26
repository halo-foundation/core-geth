#!/usr/bin/env node
/**
 * Halo Chain MAX TPS Stress Test
 * Verifies the theoretical 7,142 TPS limit
 *
 * Strategy:
 * 1. Use multiple wallets to avoid per-account limits (128 slots)
 * 2. Keep mempool saturated with transactions
 * 3. Measure actual TPS over sustained period
 */

const { ethers } = require('ethers');
const fs = require('fs');

const RPC_URL = 'http://localhost:8545';
const TARGET_TPS = 7142;
const TEST_DURATION_SECONDS = 60; // Run for 60 seconds
const REFILL_THRESHOLD = 4000; // Refill when mempool drops below this

async function main() {
    console.log('⚡ Halo Chain MAX TPS Stress Test');
    console.log('='.repeat(70));
    console.log('');
    console.log(`🎯 Target: ${TARGET_TPS} TPS`);
    console.log(`⏱️  Duration: ${TEST_DURATION_SECONDS} seconds`);
    console.log('');

    const provider = new ethers.JsonRpcProvider(RPC_URL);

    // Load test wallets
    let wallets;
    try {
        const walletsData = JSON.parse(fs.readFileSync('/tmp/test-wallets.json', 'utf8'));
        wallets = walletsData.map(w => new ethers.Wallet(w.privateKey, provider));
        console.log(`✅ Loaded ${wallets.length} test wallets`);
    } catch (error) {
        console.error('❌ Cannot load test wallets. Run: npm run export-keys');
        process.exit(1);
    }

    // Verify we have enough wallets
    const requiredWallets = Math.ceil(TARGET_TPS / 100); // Conservative estimate
    if (wallets.length < requiredWallets) {
        console.log(`⚠️  Warning: ${wallets.length} wallets may not be enough for ${TARGET_TPS} TPS`);
        console.log(`   Recommended: ${requiredWallets}+ wallets`);
    }

    const network = await provider.getNetwork();
    const startBlock = await provider.getBlockNumber();

    console.log(`📡 Chain ID: ${network.chainId}`);
    console.log(`📊 Start Block: ${startBlock}`);
    console.log('');

    // Get starting nonces
    const nonces = await Promise.all(
        wallets.map(w => provider.getTransactionCount(w.address))
    );

    console.log('🚀 Starting MAX TPS stress test...');
    console.log('   Strategy: Keep mempool saturated with transactions');
    console.log('');

    const testStartTime = Date.now();
    const testEndTime = testStartTime + (TEST_DURATION_SECONDS * 1000);

    let totalSent = 0;
    let totalErrors = 0;
    const allTxHashes = [];

    // Phase 1: Initial saturation - fill the mempool
    console.log('📤 Phase 1: Saturating mempool...');

    const initialBatch = 8000; // Fill to global slots limit
    const batchPromises = [];

    for (let i = 0; i < initialBatch; i++) {
        const walletIndex = i % wallets.length;
        const wallet = wallets[walletIndex];
        const recipient = wallets[(walletIndex + 1) % wallets.length].address;

        const txPromise = wallet.sendTransaction({
            to: recipient,
            value: ethers.parseEther('0.001'),
            gasLimit: 21000,
            nonce: nonces[walletIndex]++
        }).then(tx => {
            allTxHashes.push(tx.hash);
            totalSent++;
            if (totalSent % 1000 === 0) {
                process.stdout.write(`\r   Sent: ${totalSent}...`);
            }
            return tx;
        }).catch(err => {
            totalErrors++;
            return null;
        });

        batchPromises.push(txPromise);

        // Send in mini-batches to avoid overwhelming the RPC
        if (batchPromises.length >= 100) {
            await Promise.all(batchPromises);
            batchPromises.length = 0;
        }
    }

    // Wait for remaining
    if (batchPromises.length > 0) {
        await Promise.all(batchPromises);
    }

    console.log('');
    console.log(`✅ Mempool saturated with ${totalSent} transactions`);
    console.log('');

    // Phase 2: Continuous refill
    console.log('🔄 Phase 2: Continuous operation...');
    console.log('');

    let lastCheck = Date.now();
    let lastBlock = startBlock;

    while (Date.now() < testEndTime) {
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Check how many are still pending
        const currentBlock = await provider.getBlockNumber();
        const blocksDelta = currentBlock - lastBlock;

        if (blocksDelta > 0) {
            // Estimate how many were mined
            const txsMined = blocksDelta * 1000; // Rough estimate

            // Refill
            const toSend = Math.min(txsMined, REFILL_THRESHOLD);

            for (let i = 0; i < toSend && Date.now() < testEndTime; i++) {
                const walletIndex = totalSent % wallets.length;
                const wallet = wallets[walletIndex];
                const recipient = wallets[(walletIndex + 1) % wallets.length].address;

                wallet.sendTransaction({
                    to: recipient,
                    value: ethers.parseEther('0.001'),
                    gasLimit: 21000,
                    nonce: nonces[walletIndex]++
                }).then(tx => {
                    allTxHashes.push(tx.hash);
                    totalSent++;
                }).catch(() => {
                    totalErrors++;
                });
            }

            lastBlock = currentBlock;
        }

        const elapsed = (Date.now() - testStartTime) / 1000;
        const remaining = TEST_DURATION_SECONDS - elapsed;
        process.stdout.write(`\r   Running: ${elapsed.toFixed(0)}s | Sent: ${totalSent} | Errors: ${totalErrors} | ${remaining.toFixed(0)}s remaining   `);
    }

    console.log('');
    console.log('');
    console.log('✅ Stress test complete!');
    console.log(`   Total sent: ${totalSent.toLocaleString()}`);
    console.log(`   Errors: ${totalErrors}`);
    console.log('');

    // Phase 3: Wait for mining to complete
    console.log('⏳ Waiting for all transactions to be mined...');
    console.log('');

    const miningDeadline = Date.now() + 120000; // 2 minutes
    let mined = 0;
    let lastMined = 0;

    while (Date.now() < miningDeadline) {
        await new Promise(resolve => setTimeout(resolve, 3000));

        // Sample transactions to check mining progress
        const sampleSize = Math.min(1000, allTxHashes.length);
        const sampleIndices = [];
        for (let i = 0; i < sampleSize; i++) {
            sampleIndices.push(Math.floor(Math.random() * allTxHashes.length));
        }

        const sampleReceipts = await Promise.all(
            sampleIndices.map(i => provider.getTransactionReceipt(allTxHashes[i]).catch(() => null))
        );

        const sampleMined = sampleReceipts.filter(r => r && r.blockNumber).length;
        mined = Math.floor((sampleMined / sampleSize) * allTxHashes.length);

        if (mined > lastMined) {
            process.stdout.write(`\r   Estimated mined: ${mined.toLocaleString()}/${allTxHashes.length} (${(mined/allTxHashes.length*100).toFixed(1)}%)   `);
            lastMined = mined;
        }

        // If we've reached 95%+, we're good
        if (mined >= allTxHashes.length * 0.95) {
            break;
        }
    }

    console.log('');
    console.log('');

    const endBlock = await provider.getBlockNumber();
    const blocksProduced = endBlock - startBlock;
    const totalTime = (Date.now() - testStartTime) / 1000;

    // Calculate actual TPS
    const actualTPS = Math.floor(mined / blocksProduced); // TPS = tx per block (with 1s blocks)
    const avgTxPerBlock = Math.floor(mined / blocksProduced);

    console.log('📊 STRESS TEST RESULTS');
    console.log('='.repeat(70));
    console.log('');
    console.log('Transaction Stats:');
    console.log(`   Total Sent: ${totalSent.toLocaleString()}`);
    console.log(`   Estimated Mined: ${mined.toLocaleString()}`);
    console.log(`   Errors: ${totalErrors.toLocaleString()}`);
    console.log('');
    console.log('Block Stats:');
    console.log(`   Start Block: ${startBlock}`);
    console.log(`   End Block: ${endBlock}`);
    console.log(`   Blocks Produced: ${blocksProduced}`);
    console.log(`   Avg Tx per Block: ${avgTxPerBlock.toLocaleString()}`);
    console.log('');
    console.log('='.repeat(70));
    console.log('');
    console.log('🎯 TPS ANALYSIS (Industry Standard):');
    console.log('');
    console.log(`   Block Time: 1.000s`);
    console.log(`   Actual TPS: ${actualTPS.toLocaleString()} TPS`);
    console.log(`   Target TPS: ${TARGET_TPS.toLocaleString()} TPS`);
    console.log(`   Achievement: ${(actualTPS / TARGET_TPS * 100).toFixed(1)}%`);
    console.log('');
    console.log(`   Formula: ${avgTxPerBlock} tx/block ÷ 1s block time = ${actualTPS} TPS`);
    console.log('');
    console.log('='.repeat(70));
    console.log('');

    // Analyze sample blocks
    console.log('📦 Block Analysis (last 20 blocks):');
    console.log('');

    const blockSamples = [];
    for (let i = 1; i <= 20 && i <= blocksProduced; i++) {
        const block = await provider.getBlock(endBlock - i);
        if (block) {
            blockSamples.push({
                number: block.number,
                txCount: block.transactions.length,
                gasUsed: Number(block.gasUsed),
                gasLimit: Number(block.gasLimit)
            });
        }
    }

    blockSamples.reverse().forEach(b => {
        const utilization = (b.gasUsed / b.gasLimit * 100).toFixed(1);
        console.log(`   Block ${b.number}: ${b.txCount.toLocaleString()} txs | Gas: ${b.gasUsed.toLocaleString()}/${b.gasLimit.toLocaleString()} (${utilization}%)`);
    });

    const avgTxCount = blockSamples.reduce((sum, b) => sum + b.txCount, 0) / blockSamples.length;
    const avgGasUtil = blockSamples.reduce((sum, b) => sum + (b.gasUsed / b.gasLimit), 0) / blockSamples.length * 100;

    console.log('');
    console.log(`   Average Tx/Block: ${avgTxCount.toFixed(0)}`);
    console.log(`   Average Gas Utilization: ${avgGasUtil.toFixed(1)}%`);
    console.log('');

    // Final verdict
    console.log('='.repeat(70));
    console.log('');

    if (actualTPS >= TARGET_TPS * 0.9) {
        console.log('🎉 SUCCESS! Halo chain achieved near-theoretical maximum TPS!');
    } else if (actualTPS >= TARGET_TPS * 0.5) {
        console.log('✅ GOOD! Halo chain achieved significant TPS.');
        console.log(`   Note: Reaching ${TARGET_TPS} TPS requires optimal conditions.`);
    } else {
        console.log('⚠️  TPS lower than expected.');
        console.log('   Possible causes:');
        console.log('   - Transaction submission rate too low');
        console.log('   - Mempool limits');
        console.log('   - Network congestion');
    }

    console.log('');
    console.log('='.repeat(70));

    // Save results
    const results = {
        timestamp: new Date().toISOString(),
        test_duration_seconds: TEST_DURATION_SECONDS,
        transactions_sent: totalSent,
        transactions_mined: mined,
        blocks_produced: blocksProduced,
        avg_tx_per_block: avgTxPerBlock,
        actual_tps: actualTPS,
        target_tps: TARGET_TPS,
        achievement_percent: (actualTPS / TARGET_TPS * 100),
        avg_gas_utilization_percent: avgGasUtil
    };

    fs.writeFileSync('/tmp/stress-test-max-tps-results.json', JSON.stringify(results, null, 2));
    console.log('');
    console.log('💾 Results saved to /tmp/stress-test-max-tps-results.json');
    console.log('');
}

main().catch(console.error);
