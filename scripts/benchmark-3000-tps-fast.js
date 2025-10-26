#!/usr/bin/env node
/**
 * Halo Chain 3000 TPS Benchmark - FAST VERSION
 * Sends all transactions simultaneously (industry standard)
 */

const { ethers } = require('ethers');
const fs = require('fs');

const RPC_URL = 'http://localhost:8545';
const TOTAL_TXS = 30000; // 30,000 transactions

async function main() {
    console.log('⚡ Halo Chain 3000 TPS Benchmark (FAST)');
    console.log('='.repeat(70));
    console.log(`Total Transactions: ${TOTAL_TXS.toLocaleString()}`);
    console.log('Strategy: Send all transactions simultaneously');
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

    // Get starting nonces
    const startNonces = await Promise.all(
        wallets.map(w => provider.getTransactionCount(w.address))
    );

    console.log('🚀 Pre-signing and sending ALL transactions simultaneously...');
    console.log('');

    const startTime = Date.now();
    const txsPerWallet = Math.ceil(TOTAL_TXS / wallets.length);

    // Build all transaction promises
    const allPromises = [];

    for (let walletIndex = 0; walletIndex < wallets.length; walletIndex++) {
        const wallet = wallets[walletIndex];
        const recipient = wallets[(walletIndex + 1) % wallets.length].address;
        const nonce = startNonces[walletIndex];

        for (let i = 0; i < txsPerWallet; i++) {
            if (allPromises.length >= TOTAL_TXS) break;

            // Create transaction promise (don't await!)
            const txPromise = wallet.sendTransaction({
                to: recipient,
                value: ethers.parseEther('0.001'),
                gasLimit: 21000,
                nonce: nonce + i
            }).catch(err => ({ error: err.message }));

            allPromises.push(txPromise);
        }
    }

    console.log(`📤 Firing ${allPromises.length} transactions...`);

    // Send ALL transactions at once!
    const results = await Promise.all(allPromises);

    const sendEndTime = Date.now();
    const sendDuration = (sendEndTime - startTime) / 1000;

    // Count successes
    const successful = results.filter(r => !r.error);
    const errors = results.filter(r => r.error);

    console.log('');
    console.log('✅ All transactions submitted!');
    console.log(`   Total Submitted: ${results.length.toLocaleString()}`);
    console.log(`   Successful: ${successful.length.toLocaleString()}`);
    console.log(`   Errors: ${errors.length}`);
    console.log(`   Send Duration: ${sendDuration.toFixed(2)}s`);
    console.log(`   Send Rate: ${(successful.length / sendDuration).toFixed(0)} TPS`);
    console.log('');

    // Sample errors
    if (errors.length > 0 && errors.length <= 5) {
        console.log('Error samples:');
        errors.forEach((e, i) => {
            console.log(`   ${i+1}. ${e.error.substring(0, 80)}`);
        });
        console.log('');
    }

    // Wait for mining
    console.log('⏳ Waiting for transactions to be mined...');
    console.log('');

    // Get tx hashes
    const txHashes = successful.map(r => r.hash).filter(h => h);

    let mined = 0;
    const startMiningTime = Date.now();
    const timeout = 120000; // 2 minutes

    while (mined < txHashes.length && (Date.now() - startMiningTime) < timeout) {
        await new Promise(resolve => setTimeout(resolve, 3000));

        // Check how many are mined
        const receipts = await Promise.all(
            txHashes.map(h => provider.getTransactionReceipt(h).catch(() => null))
        );

        const newMined = receipts.filter(r => r && r.blockNumber).length;

        if (newMined > mined) {
            mined = newMined;
            const elapsed = (Date.now() - startMiningTime) / 1000;
            const miningTPS = mined / elapsed;
            console.log(`   Mined: ${mined.toLocaleString()}/${txHashes.length} (${miningTPS.toFixed(0)} TPS) [${txHashes.length - mined} pending]`);
        }
    }

    const endBlock = await provider.getBlockNumber();
    const totalDuration = (Date.now() - startTime) / 1000;

    console.log('');
    console.log('📊 BENCHMARK RESULTS');
    console.log('='.repeat(70));
    console.log('');
    console.log('Transaction Stats:');
    console.log(`   Total Submitted: ${successful.length.toLocaleString()}`);
    console.log(`   Total Mined: ${mined.toLocaleString()}`);
    console.log(`   Success Rate: ${((mined / successful.length) * 100).toFixed(2)}%`);
    console.log(`   Errors: ${errors.length}`);
    console.log('');
    console.log('Performance:');
    console.log(`   Send Rate: ${(successful.length / sendDuration).toFixed(0)} TPS`);
    console.log(`   Mining Rate: ${(mined / totalDuration).toFixed(0)} TPS`);
    console.log(`   Target TPS: 3000`);
    console.log(`   Achievement: ${((mined / totalDuration / 3000) * 100).toFixed(2)}%`);
    console.log('');
    console.log('Block Info:');
    console.log(`   Start Block: ${startBlock}`);
    console.log(`   End Block: ${endBlock}`);
    console.log(`   Blocks Produced: ${endBlock - startBlock}`);
    console.log(`   Avg TPS per Block: ${(mined / (endBlock - startBlock)).toFixed(0)}`);
    console.log('');

    // Analyze blocks
    console.log('⏱️  Analyzing blocks with transactions...');
    const blockNumbers = new Set();

    // Get unique block numbers
    const receipts = await Promise.all(
        txHashes.slice(0, Math.min(txHashes.length, 100)).map(h =>
            provider.getTransactionReceipt(h).catch(() => null)
        )
    );

    receipts.filter(r => r).forEach(r => blockNumbers.add(Number(r.blockNumber)));

    // Analyze sample blocks
    const sampleBlocks = Array.from(blockNumbers).slice(0, 10);
    console.log('');
    console.log(`Sample Blocks (first 10 of ${blockNumbers.size}):`);

    for (const blockNum of sampleBlocks) {
        const block = await provider.getBlock(blockNum);
        const prevBlock = await provider.getBlock(blockNum - 1);

        if (block && prevBlock) {
            const blockTime = Number(block.timestamp) - Number(prevBlock.timestamp);
            const txCount = block.transactions.length;
            const gasUsed = Number(block.gasUsed);
            const gasLimit = Number(block.gasLimit);
            const tps = txCount / blockTime;

            console.log(`   Block ${blockNum}: ${txCount} txs in ${blockTime}s = ${tps.toFixed(0)} TPS (gas: ${gasUsed.toLocaleString()}/${gasLimit.toLocaleString()})`);
        }
    }

    // Check gas limit
    const latestBlock = await provider.getBlock(endBlock);
    console.log('');
    console.log('Gas Limit Verification:');
    console.log(`   Current Gas Limit: ${Number(latestBlock.gasLimit).toLocaleString()}`);
    console.log(`   Expected: 150,000,000`);
    console.log(`   Match: ${latestBlock.gasLimit.toString() === '150000000' ? '✅' : '❌'}`);

    console.log('');
    console.log('='.repeat(70));

    const actualTPS = mined / totalDuration;
    if (actualTPS >= 3000) {
        console.log('🎉 SUCCESS: Achieved 3000+ TPS target!');
    } else {
        console.log(`⚠️  Target not reached: ${actualTPS.toFixed(0)} TPS (need 3000 TPS)`);
    }

    console.log('='.repeat(70));

    // Save results
    const results_data = {
        timestamp: new Date().toISOString(),
        totalSubmitted: successful.length,
        totalMined: mined,
        sendDuration: sendDuration,
        totalDuration: totalDuration,
        sendRate: successful.length / sendDuration,
        miningRate: mined / totalDuration,
        successRate: (mined / successful.length) * 100,
        errors: errors.length,
        startBlock: startBlock,
        endBlock: endBlock,
        blocksProduced: endBlock - startBlock,
        gasLimit: latestBlock.gasLimit.toString()
    };

    fs.writeFileSync('/tmp/tps-benchmark-fast-results.json', JSON.stringify(results_data, null, 2));
    console.log('');
    console.log('💾 Results saved to /tmp/tps-benchmark-fast-results.json');
}

main().catch(console.error);
