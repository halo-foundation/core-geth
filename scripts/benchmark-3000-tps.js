#!/usr/bin/env node
/**
 * Halo Chain 3000 TPS Benchmark Test
 * Target: 3000 transactions per second
 * Gas Limit: 150M per block
 * Block Time: 1 second
 */

const { ethers } = require('ethers');
const fs = require('fs');

const RPC_URL = 'http://localhost:8545';
const TEST_DURATION_SECONDS = 10; // Run for 10 seconds
const TARGET_TPS = 3000;
const TOTAL_TXS = TARGET_TPS * TEST_DURATION_SECONDS; // 30,000 transactions

async function main() {
    console.log('⚡ Halo Chain 3000 TPS Benchmark');
    console.log('='.repeat(70));
    console.log(`Target TPS: ${TARGET_TPS}`);
    console.log(`Test Duration: ${TEST_DURATION_SECONDS} seconds`);
    console.log(`Total Transactions: ${TOTAL_TXS.toLocaleString()}`);
    console.log('');

    const provider = new ethers.JsonRpcProvider(RPC_URL);

    // Load test wallets
    const walletsData = JSON.parse(fs.readFileSync('/tmp/test-wallets.json', 'utf8'));
    const wallets = walletsData.map(w => new ethers.Wallet(w.privateKey, provider));

    console.log(`✅ Loaded ${wallets.length} test wallets`);

    // Check network
    const network = await provider.getNetwork();
    const startBlock = await provider.getBlockNumber();
    console.log(`📡 Connected to Chain ID: ${network.chainId}`);
    console.log(`📊 Starting Block: ${startBlock}`);
    console.log('');

    // Get starting nonces for all wallets
    const startNonces = [];
    for (let i = 0; i < wallets.length; i++) {
        const nonce = await provider.getTransactionCount(wallets[i].address);
        startNonces.push(nonce);
    }

    console.log('🚀 Starting benchmark...');
    console.log('');

    const startTime = Date.now();
    const txHashes = [];
    let txsSent = 0;
    let errors = 0;

    // Send transactions in parallel from all wallets
    const txsPerWallet = Math.ceil(TOTAL_TXS / wallets.length);

    const promises = wallets.map(async (wallet, walletIndex) => {
        const nonce = startNonces[walletIndex];
        const recipient = wallets[(walletIndex + 1) % wallets.length].address;

        for (let i = 0; i < txsPerWallet && txsSent < TOTAL_TXS; i++) {
            try {
                const tx = await wallet.sendTransaction({
                    to: recipient,
                    value: ethers.parseEther('0.001'),
                    gasLimit: 21000,
                    nonce: nonce + i
                });

                txHashes.push(tx.hash);
                txsSent++;

                if (txsSent % 1000 === 0) {
                    const elapsed = (Date.now() - startTime) / 1000;
                    const currentTPS = txsSent / elapsed;
                    console.log(`   Sent: ${txsSent.toLocaleString()} txs (${currentTPS.toFixed(0)} TPS)`);
                }
            } catch (e) {
                errors++;
                if (errors <= 5) {
                    console.log(`   ⚠️  Error: ${e.message.substring(0, 80)}`);
                }
            }
        }
    });

    await Promise.all(promises);

    const sendEndTime = Date.now();
    const sendDuration = (sendEndTime - startTime) / 1000;

    console.log('');
    console.log('✅ All transactions sent!');
    console.log(`   Total Sent: ${txsSent.toLocaleString()}`);
    console.log(`   Errors: ${errors}`);
    console.log(`   Send Duration: ${sendDuration.toFixed(2)}s`);
    console.log(`   Send Rate: ${(txsSent / sendDuration).toFixed(0)} TPS`);
    console.log('');

    // Wait for transactions to be mined
    console.log('⏳ Waiting for transactions to be mined...');
    console.log('');

    let mined = 0;
    let pending = txHashes.length;
    const checkInterval = 2000; // Check every 2 seconds

    const startMiningTime = Date.now();

    while (pending > 0 && (Date.now() - startMiningTime) < 120000) { // 2 minute timeout
        await new Promise(resolve => setTimeout(resolve, checkInterval));

        let newlyMined = 0;
        for (const hash of txHashes) {
            try {
                const receipt = await provider.getTransactionReceipt(hash);
                if (receipt && receipt.blockNumber) {
                    newlyMined++;
                }
            } catch (e) {
                // Transaction not yet mined
            }
        }

        if (newlyMined > mined) {
            mined = newlyMined;
            pending = txHashes.length - mined;
            const elapsed = (Date.now() - startMiningTime) / 1000;
            const miningTPS = mined / elapsed;
            console.log(`   Mined: ${mined.toLocaleString()}/${txHashes.length} (${miningTPS.toFixed(0)} TPS) [${pending} pending]`);
        }
    }

    const endBlock = await provider.getBlockNumber();
    const totalDuration = (Date.now() - startTime) / 1000;

    console.log('');
    console.log('📊 BENCHMARK RESULTS');
    console.log('='.repeat(70));
    console.log('');
    console.log('Transaction Stats:');
    console.log(`   Total Sent: ${txsSent.toLocaleString()}`);
    console.log(`   Total Mined: ${mined.toLocaleString()}`);
    console.log(`   Success Rate: ${((mined / txsSent) * 100).toFixed(2)}%`);
    console.log(`   Errors: ${errors}`);
    console.log('');
    console.log('Performance:');
    console.log(`   Send Rate: ${(txsSent / sendDuration).toFixed(0)} TPS`);
    console.log(`   Mining Rate: ${(mined / totalDuration).toFixed(0)} TPS`);
    console.log(`   Target TPS: ${TARGET_TPS}`);
    console.log(`   Achievement: ${((mined / totalDuration / TARGET_TPS) * 100).toFixed(2)}%`);
    console.log('');
    console.log('Block Info:');
    console.log(`   Start Block: ${startBlock}`);
    console.log(`   End Block: ${endBlock}`);
    console.log(`   Blocks Produced: ${endBlock - startBlock}`);
    console.log(`   Avg TPS per Block: ${(mined / (endBlock - startBlock)).toFixed(0)}`);
    console.log('');

    // Analyze block times
    console.log('⏱️  Analyzing block times...');
    const blockTimes = [];
    for (let i = startBlock + 1; i <= endBlock; i++) {
        const block = await provider.getBlock(i);
        const prevBlock = await provider.getBlock(i - 1);
        if (block && prevBlock) {
            blockTimes.push(Number(block.timestamp) - Number(prevBlock.timestamp));
        }
    }

    if (blockTimes.length > 0) {
        const avgBlockTime = blockTimes.reduce((a, b) => a + b, 0) / blockTimes.length;
        const minBlockTime = Math.min(...blockTimes);
        const maxBlockTime = Math.max(...blockTimes);

        console.log('');
        console.log('Block Time Analysis:');
        console.log(`   Average: ${avgBlockTime.toFixed(2)}s`);
        console.log(`   Min: ${minBlockTime}s`);
        console.log(`   Max: ${maxBlockTime}s`);
        console.log(`   Target: 1s`);
        console.log(`   Consistency: ${blockTimes.filter(t => t === 1).length}/${blockTimes.length} blocks at 1s`);
    }

    // Check gas limit
    const latestBlock = await provider.getBlock(endBlock);
    console.log('');
    console.log('Gas Limit Verification:');
    console.log(`   Current Gas Limit: ${latestBlock.gasLimit.toLocaleString()}`);
    console.log(`   Expected: 150,000,000`);
    console.log(`   Match: ${latestBlock.gasLimit.toString() === '150000000' ? '✅' : '❌'}`);

    console.log('');
    console.log('='.repeat(70));

    if (mined / totalDuration >= TARGET_TPS) {
        console.log('🎉 SUCCESS: Achieved 3000+ TPS target!');
    } else {
        console.log(`⚠️  Target not reached: ${(mined / totalDuration).toFixed(0)} TPS (need ${TARGET_TPS} TPS)`);
    }

    console.log('='.repeat(70));

    // Save results
    const results = {
        timestamp: new Date().toISOString(),
        targetTPS: TARGET_TPS,
        txsSent: txsSent,
        txsMined: mined,
        sendDuration: sendDuration,
        totalDuration: totalDuration,
        sendRate: txsSent / sendDuration,
        miningRate: mined / totalDuration,
        successRate: (mined / txsSent) * 100,
        errors: errors,
        startBlock: startBlock,
        endBlock: endBlock,
        blocksProduced: endBlock - startBlock,
        blockTimes: blockTimes,
        avgBlockTime: blockTimes.reduce((a, b) => a + b, 0) / blockTimes.length,
        gasLimit: latestBlock.gasLimit.toString()
    };

    fs.writeFileSync('/tmp/tps-benchmark-results.json', JSON.stringify(results, null, 2));
    console.log('');
    console.log('💾 Results saved to /tmp/tps-benchmark-results.json');
}

main().catch(console.error);
