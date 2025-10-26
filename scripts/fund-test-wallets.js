#!/usr/bin/env node
/**
 * Fund test wallets for TPS benchmark
 */

const { ethers } = require('ethers');

const RPC_URL = 'http://localhost:8545';
const MINER_ADDR = '0x69AEd36e548525ED741052A6572Bb1328973b44F';
const MINER_KEY = '0x32e1b0aeb11846cc691c407821280d5c78be0249c7c9746cd3e81e81ea2e937e';

// Generate 10 test wallets
const NUM_WALLETS = 10;
const FUND_AMOUNT = ethers.parseEther('100'); // 100 HALO each

async function main() {
    console.log('💰 Funding Test Wallets for TPS Benchmark');
    console.log('='.repeat(70));
    console.log('');

    const provider = new ethers.JsonRpcProvider(RPC_URL);

    // Create miner wallet from private key
    const minerWallet = new ethers.Wallet(MINER_KEY, provider);
    console.log(`✅ Loaded miner wallet: ${minerWallet.address}`);

    const minerBalance = await provider.getBalance(minerWallet.address);
    console.log(`   Balance: ${ethers.formatEther(minerBalance)} HALO`);
    console.log('');

    // Create random wallets
    const wallets = [];
    for (let i = 0; i < NUM_WALLETS; i++) {
        const wallet = ethers.Wallet.createRandom().connect(provider);
        wallets.push(wallet);
    }

    console.log(`✅ Generated ${NUM_WALLETS} test wallets`);
    console.log('');

    // Fund each wallet with proper nonce management
    console.log('💸 Funding wallets...');
    let nonce = await provider.getTransactionCount(minerWallet.address);

    const txHashes = [];
    for (let i = 0; i < wallets.length; i++) {
        const tx = await minerWallet.sendTransaction({
            to: wallets[i].address,
            value: FUND_AMOUNT,
            gasLimit: 21000,
            nonce: nonce + i
        });
        txHashes.push(tx.hash);
        console.log(`   Wallet ${i+1}: ${wallets[i].address} (tx: ${tx.hash})`);
    }

    // Wait for all transactions to be mined
    console.log('');
    console.log('⏳ Waiting for all transactions to be mined...');
    for (const hash of txHashes) {
        await provider.waitForTransaction(hash);
    }

    console.log('');
    console.log('⏳ Waiting for confirmations...');

    // Wait a few seconds for mining
    await new Promise(resolve => setTimeout(resolve, 5000));

    // Verify balances
    console.log('');
    console.log('✅ Funded Wallets:');
    for (let i = 0; i < wallets.length; i++) {
        const balance = await provider.getBalance(wallets[i].address);
        console.log(`   ${wallets[i].address}: ${ethers.formatEther(balance)} HALO`);
    }

    // Save wallet info
    const walletsData = wallets.map(w => ({
        address: w.address,
        privateKey: w.privateKey
    }));

    const fs = require('fs');
    fs.writeFileSync('/tmp/test-wallets.json', JSON.stringify(walletsData, null, 2));
    console.log('');
    console.log('💾 Wallet data saved to /tmp/test-wallets.json');
}

main().catch(console.error);
