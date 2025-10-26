#!/usr/bin/env node
/**
 * Verify Halo Chain Block Time
 */

const { ethers } = require('ethers');

const RPC_URL = 'http://localhost:8545';
const BLOCKS_TO_CHECK = 100;

async function main() {
    console.log('⏱️  Verifying Halo Chain Block Time');
    console.log('='.repeat(70));
    console.log('');

    const provider = new ethers.JsonRpcProvider(RPC_URL);

    const currentBlock = await provider.getBlockNumber();
    console.log(`Current Block: ${currentBlock}`);
    console.log(`Analyzing last ${BLOCKS_TO_CHECK} blocks...`);
    console.log('');

    const blockTimes = [];
    let totalTime = 0;

    for (let i = 0; i < BLOCKS_TO_CHECK && i < currentBlock; i++) {
        const blockNum = currentBlock - i;
        const block = await provider.getBlock(blockNum);
        const prevBlock = await provider.getBlock(blockNum - 1);

        if (block && prevBlock) {
            const blockTime = Number(block.timestamp) - Number(prevBlock.timestamp);
            blockTimes.push(blockTime);
            totalTime += blockTime;

            if (i < 10) {
                console.log(`Block ${blockNum}: ${blockTime}s (${block.transactions.length} txs)`);
            }
        }
    }

    console.log('');
    console.log('📊 Block Time Analysis:');
    console.log(`   Blocks Analyzed: ${blockTimes.length}`);
    console.log(`   Average Block Time: ${(totalTime / blockTimes.length).toFixed(3)}s`);
    console.log(`   Min Block Time: ${Math.min(...blockTimes)}s`);
    console.log(`   Max Block Time: ${Math.max(...blockTimes)}s`);
    console.log('');

    // Check if it's exactly 1 second
    const avgBlockTime = totalTime / blockTimes.length;
    const target = 1.0;
    const deviation = Math.abs(avgBlockTime - target);

    console.log('🎯 Target Verification:');
    console.log(`   Target: ${target}s`);
    console.log(`   Actual: ${avgBlockTime.toFixed(3)}s`);
    console.log(`   Deviation: ${deviation.toFixed(3)}s`);
    console.log(`   Status: ${deviation < 0.01 ? '✅ EXACT' : deviation < 0.1 ? '✅ GOOD' : '⚠️  OFF TARGET'}`);
    console.log('');

    // Distribution
    const oneSecBlocks = blockTimes.filter(t => t === 1).length;
    const lessThanOne = blockTimes.filter(t => t < 1).length;
    const moreThanOne = blockTimes.filter(t => t > 1).length;

    console.log('📈 Distribution:');
    console.log(`   Exactly 1s: ${oneSecBlocks} (${(oneSecBlocks/blockTimes.length*100).toFixed(1)}%)`);
    console.log(`   < 1s: ${lessThanOne} (${(lessThanOne/blockTimes.length*100).toFixed(1)}%)`);
    console.log(`   > 1s: ${moreThanOne} (${(moreThanOne/blockTimes.length*100).toFixed(1)}%)`);
    console.log('');

    // Calculate realistic TPS
    const latestBlock = await provider.getBlock(currentBlock);
    const gasLimit = Number(latestBlock.gasLimit);
    const maxTxPerBlock = Math.floor(gasLimit / 21000);
    const theoreticalTPS = maxTxPerBlock / avgBlockTime;

    console.log('🚀 TPS Capacity:');
    console.log(`   Gas Limit: ${gasLimit.toLocaleString()}`);
    console.log(`   Max Tx/Block: ${maxTxPerBlock.toLocaleString()} (simple transfers)`);
    console.log(`   Block Time: ${avgBlockTime.toFixed(3)}s`);
    console.log(`   Theoretical Max TPS: ${theoreticalTPS.toFixed(0)} TPS`);
    console.log('');
}

main().catch(console.error);
