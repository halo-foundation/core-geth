#!/usr/bin/env node
/**
 * CORRECT Halo Chain TPS Benchmark
 * Uses industry-standard TPS calculation: Total Transactions / Total Seconds
 */

const { ethers } = require('ethers');
const fs = require('fs');

const RPC_URL = 'http://localhost:8545';
const TEST_DURATION_SECONDS = 30; // Run for 30 seconds
const MAX_TXS_TO_SEND = 10000; // Cap to prevent runaway

async function main() {
    console.log('⚡ CORRECT Halo Chain TPS Benchmark');
    console.log('='.repeat(70));
    console.log('Using Industry Standard: TPS = Total Transactions / Total Seconds');
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

    // Check network
    const network = await provider.getNetwork();
    const startBlock = await provider.getBlockNumber();
    console.log(`📡 Connected to Chain ID: ${network.chainId}`);
    console.log(`📊 Starting Block: ${startBlock}`);
    console.log('');

    // Get starting nonces
    const startNonces = await Promise.all(
        wallets.map(w => provider.getTransactionCount(w.address))
    );

    console.log(`⚙️  Configuration:`);
    console.log(`   Test Duration: ${TEST_DURATION_SECONDS} seconds`);
    console.log(`   Wallets: ${wallets.length}`);
    console.log(`   Max Transactions: ${MAX_TXS_TO_SEND.toLocaleString()}`);
    console.log('');

    console.log('🚀 Starting TPS benchmark...');
    console.log('');

    const startTime = Date.now();
    const endTime = startTime + (TEST_DURATION_SECONDS * 1000);

    let txCount = 0;
    let walletIndex = 0;
    const txHashes = [];
    const nonces = [...startNonces];

    // Send transactions continuously for the duration
    while (Date.now() < endTime && txCount < MAX_TXS_TO_SEND) {
        const wallet = wallets[walletIndex];
        const recipient = wallets[(walletIndex + 1) % wallets.length].address;

        try {
            const tx = await wallet.sendTransaction({
                to: recipient,
                value: ethers.parseEther('0.001'),
                gasLimit: 21000,
                nonce: nonces[walletIndex]++
            });

            txHashes.push(tx.hash);
            txCount++;

            if (txCount % 100 === 0) {
                const elapsed = (Date.now() - startTime) / 1000;
                const currentRate = txCount / elapsed;
                process.stdout.write(`\r   Sent: ${txCount} txs | ${currentRate.toFixed(0)} tx/s | ${(TEST_DURATION_SECONDS - elapsed).toFixed(0)}s remaining   `);
            }
        } catch (error) {
            // Skip errors and continue
        }

        // Rotate to next wallet
        walletIndex = (walletIndex + 1) % wallets.length;
    }

    const sendEndTime = Date.now();
    const sendDuration = (sendEndTime - startTime) / 1000;

    console.log('');
    console.log('');
    console.log(`✅ Sent ${txCount.toLocaleString()} transactions in ${sendDuration.toFixed(2)}s`);
    console.log(`   Send Rate: ${(txCount / sendDuration).toFixed(0)} tx/s`);
    console.log('');

    // Wait for mining (give it time equal to send duration)
    console.log('⏳ Waiting for transactions to be mined...');
    const miningTimeout = Math.max(sendDuration * 1000, 60000); // At least 60s
    const miningDeadline = Date.now() + miningTimeout;

    let mined = 0;
    let lastMined = 0;
    const miningStartTime = Date.now();

    while (Date.now() < miningDeadline) {
        await new Promise(resolve => setTimeout(resolve, 2000));

        // Check mined count
        const receipts = await Promise.all(
            txHashes.map(h => provider.getTransactionReceipt(h).catch(() => null))
        );

        mined = receipts.filter(r => r && r.blockNumber).length;

        if (mined > lastMined) {
            const elapsed = (Date.now() - miningStartTime) / 1000;
            const miningRate = mined / elapsed;
            process.stdout.write(`\r   Mined: ${mined.toLocaleString()}/${txCount} (${miningRate.toFixed(0)} TPS) - ${(mined/txCount*100).toFixed(1)}% complete   `);
            lastMined = mined;
        }

        // If all mined, break
        if (mined >= txCount) {
            break;
        }
    }

    console.log('');
    console.log('');

    const endBlock = await provider.getBlockNumber();
    const totalTime = (Date.now() - startTime) / 1000;
    const miningTime = (Date.now() - miningStartTime) / 1000;

    // Calculate CORRECT TPS
    const actualTPS = mined / miningTime;
    const blocksProduced = endBlock - startBlock;
    const avgBlockTime = blocksProduced > 0 ? miningTime / blocksProduced : 0;

    console.log('📊 BENCHMARK RESULTS');
    console.log('='.repeat(70));
    console.log('');
    console.log('Transaction Stats:');
    console.log(`   Total Sent: ${txCount.toLocaleString()}`);
    console.log(`   Total Mined: ${mined.toLocaleString()}`);
    console.log(`   Success Rate: ${((mined / txCount) * 100).toFixed(2)}%`);
    console.log('');
    console.log('Timing:');
    console.log(`   Send Duration: ${sendDuration.toFixed(2)}s`);
    console.log(`   Mining Duration: ${miningTime.toFixed(2)}s`);
    console.log(`   Total Duration: ${totalTime.toFixed(2)}s`);
    console.log('');
    console.log('Block Stats:');
    console.log(`   Start Block: ${startBlock}`);
    console.log(`   End Block: ${endBlock}`);
    console.log(`   Blocks Produced: ${blocksProduced}`);
    console.log(`   Avg Block Time: ${avgBlockTime.toFixed(2)}s`);
    console.log(`   Avg Tx per Block: ${(mined / blocksProduced).toFixed(0)}`);
    console.log('');
    console.log('='.repeat(70));
    console.log('');
    console.log('🎯 ACTUAL TPS (Industry Standard):');
    console.log(`   ${actualTPS.toFixed(0)} TPS`);
    console.log('');
    console.log(`   Formula: ${mined} transactions / ${miningTime.toFixed(2)} seconds = ${actualTPS.toFixed(0)} TPS`);
    console.log('');
    console.log('='.repeat(70));

    // Analyze blocks
    console.log('');
    console.log('📦 Block Analysis:');

    // Sample recent blocks
    const sampleSize = Math.min(10, blocksProduced);
    const sampleBlocks = [];
    for (let i = 0; i < sampleSize; i++) {
        const blockNum = endBlock - i;
        const block = await provider.getBlock(blockNum);
        if (block) {
            sampleBlocks.push(block);
        }
    }

    console.log('');
    console.log(`Recent Blocks (last ${sampleBlocks.length}):`);
    for (const block of sampleBlocks.reverse()) {
        const txCount = block.transactions.length;
        const gasUsed = Number(block.gasUsed);
        const gasLimit = Number(block.gasLimit);
        const gasUtilization = (gasUsed / gasLimit * 100).toFixed(1);

        console.log(`   Block ${block.number}: ${txCount} txs | Gas: ${gasUsed.toLocaleString()}/${gasLimit.toLocaleString()} (${gasUtilization}%)`);
    }

    // Check gas limit
    const latestBlock = await provider.getBlock(endBlock);
    const currentGasLimit = Number(latestBlock.gasLimit);

    console.log('');
    console.log('⛽ Gas Limit Analysis:');
    console.log(`   Current: ${currentGasLimit.toLocaleString()}`);
    console.log(`   Expected: 150,000,000`);
    console.log(`   Match: ${currentGasLimit === 150000000 ? '✅' : '❌'}`);

    const maxTxPerBlock = Math.floor(currentGasLimit / 21000);
    const theoreticalTPS = avgBlockTime > 0 ? maxTxPerBlock / avgBlockTime : 0;

    console.log('');
    console.log('📈 Theoretical Limits:');
    console.log(`   Max Tx per Block: ${maxTxPerBlock.toLocaleString()} (at 21k gas each)`);
    console.log(`   Block Time: ${avgBlockTime.toFixed(2)}s`);
    console.log(`   Theoretical Max TPS: ${theoreticalTPS.toFixed(0)} TPS`);
    console.log(`   Current TPS: ${actualTPS.toFixed(0)} TPS`);
    console.log(`   Utilization: ${(actualTPS / theoreticalTPS * 100).toFixed(2)}%`);
    console.log('');

    // Save results
    const results = {
        timestamp: new Date().toISOString(),
        methodology: "Industry Standard: TPS = Total Transactions / Total Seconds",
        configuration: {
            test_duration_seconds: TEST_DURATION_SECONDS,
            wallets_used: wallets.length
        },
        results: {
            transactions_sent: txCount,
            transactions_mined: mined,
            success_rate_percent: (mined / txCount) * 100,
            send_duration_seconds: sendDuration,
            mining_duration_seconds: miningTime,
            total_duration_seconds: totalTime
        },
        tps: {
            actual_tps: actualTPS,
            formula: `${mined} transactions / ${miningTime.toFixed(2)} seconds`,
            theoretical_max_tps: theoreticalTPS,
            utilization_percent: (actualTPS / theoreticalTPS) * 100
        },
        blocks: {
            start_block: startBlock,
            end_block: endBlock,
            blocks_produced: blocksProduced,
            avg_block_time_seconds: avgBlockTime,
            avg_tx_per_block: mined / blocksProduced
        },
        gas: {
            gas_limit: currentGasLimit,
            max_tx_per_block: maxTxPerBlock
        }
    };

    fs.writeFileSync('/tmp/tps-correct-results.json', JSON.stringify(results, null, 2));
    console.log('💾 Results saved to /tmp/tps-correct-results.json');
    console.log('');
}

main().catch(console.error);
