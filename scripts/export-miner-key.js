#!/usr/bin/env node
/**
 * Export miner private key from keystore
 */

const fs = require('fs');
const { Wallet } = require('ethers');
const path = require('path');

const KEYSTORE_PATH = path.join(__dirname, '..', 'halo-test', 'keystore');
const PASSWORD = 'test123';

async function main() {
    try {
        // Find keystore file
        const files = fs.readdirSync(KEYSTORE_PATH);
        const keystoreFile = files.find(f => f.startsWith('UTC--'));

        if (!keystoreFile) {
            console.error('❌ No keystore file found');
            process.exit(1);
        }

        const keystorePath = path.join(KEYSTORE_PATH, keystoreFile);
        const json = fs.readFileSync(keystorePath, 'utf8');
        const wallet = await Wallet.fromEncryptedJson(json, PASSWORD);

        console.log('🔐 Miner Account');
        console.log('================');
        console.log('Address:', wallet.address);
        console.log('Private Key:', wallet.privateKey);
        console.log('');
        console.log('⚠️  Keep this private key secure!');
    } catch (error) {
        console.error('❌ Error:', error.message);
        process.exit(1);
    }
}

main();
