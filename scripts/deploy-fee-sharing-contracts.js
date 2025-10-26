// Halo Chain Fee Sharing Contract Deployment Script
// Usage: loadScript("scripts/deploy-fee-sharing-contracts.js") in geth console

console.log("==============================================");
console.log("  Halo Chain Fee Sharing Contract Deployer");
console.log("==============================================");
console.log("");

// Check if we're connected
if (!eth.syncing && eth.blockNumber >= 0) {
    console.log("✓ Connected to Halo Chain");
    console.log("  Chain ID:", eth.chainId());
    console.log("  Current Block:", eth.blockNumber);
    console.log("");
} else {
    console.log("✗ Not connected to Halo Chain or still syncing");
    console.log("  Please wait for sync to complete");
    console.log("");
}

// Configuration
var deployerAccount = eth.accounts[0];
var recipientAddress = eth.accounts[0]; // Change this to your desired recipient
var feePercent = 50; // 50% of miner fees to contract owner

console.log("Deployment Configuration:");
console.log("  Deployer:", deployerAccount);
console.log("  Fee Recipient:", recipientAddress);
console.log("  Fee Percent:", feePercent + "%");
console.log("");

// HaloFeeSharing Contract ABI (minimal for deployment)
var haloFeeSharingAbi = [
    {
        "inputs": [],
        "stateMutability": "nonpayable",
        "type": "constructor"
    },
    {
        "inputs": [{"internalType": "address","name": "recipient","type": "address"},{"internalType": "uint8","name": "percent","type": "uint8"}],
        "name": "enableFeeSharing",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "getFeeConfig",
        "outputs": [{"internalType": "bool","name": "enabled","type": "bool"},{"internalType": "address","name": "recipient","type": "address"},{"internalType": "uint8","name": "percent","type": "uint8"}],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "owner",
        "outputs": [{"internalType": "address","name": "","type": "address"}],
        "stateMutability": "view",
        "type": "function"
    }
];

// ExampleDApp Contract ABI (minimal for deployment)
var exampleDAppAbi = [
    {
        "inputs": [{"internalType": "address","name": "feeRecipient","type": "address"},{"internalType": "uint8","name": "feePercent","type": "uint8"}],
        "stateMutability": "nonpayable",
        "type": "constructor"
    },
    {
        "inputs": [{"internalType": "uint256","name": "score","type": "uint256"}],
        "name": "updateScore",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "incrementScore",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [{"internalType": "address","name": "user","type": "address"}],
        "name": "getScore",
        "outputs": [{"internalType": "uint256","name": "","type": "uint256"}],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "getFeeConfig",
        "outputs": [{"internalType": "bool","name": "enabled","type": "bool"},{"internalType": "address","name": "recipient","type": "address"},{"internalType": "uint8","name": "percent","type": "uint8"}],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "totalInteractions",
        "outputs": [{"internalType": "uint256","name": "","type": "uint256"}],
        "stateMutability": "view",
        "type": "function"
    }
];

console.log("==============================================");
console.log("STEP 1: Compile Contracts");
console.log("==============================================");
console.log("");
console.log("Before deploying, you need to compile the Solidity contracts.");
console.log("");
console.log("Option 1: Use Remix IDE");
console.log("  1. Go to https://remix.ethereum.org");
console.log("  2. Copy contracts/HaloFeeSharing.sol and contracts/ExampleDApp.sol");
console.log("  3. Compile both contracts");
console.log("  4. Copy the bytecode from the compilation details");
console.log("");
console.log("Option 2: Use solc command line");
console.log("  cd /home/blackluv/core-geth");
console.log("  solc --bin --abi contracts/HaloFeeSharing.sol -o compiled/");
console.log("  solc --bin --abi contracts/ExampleDApp.sol -o compiled/");
console.log("");
console.log("After compilation, set the bytecode variables:");
console.log("");
console.log("  var haloFeeSharingBytecode = \"0x...\";");
console.log("  var exampleDAppBytecode = \"0x...\";");
console.log("");

console.log("==============================================");
console.log("STEP 2: Deploy ExampleDApp Contract");
console.log("==============================================");
console.log("");
console.log("After setting the bytecode, run these commands:");
console.log("");
console.log("  // Unlock your account");
console.log("  personal.unlockAccount(eth.accounts[0], \"your-password\", 300);");
console.log("");
console.log("  // Create contract object");
console.log("  var ExampleDAppContract = eth.contract(exampleDAppAbi);");
console.log("");
console.log("  // Deploy the contract");
console.log("  var exampleDApp = ExampleDAppContract.new(");
console.log("    \"" + recipientAddress + "\", // Fee recipient");
console.log("    " + feePercent + ",        // Fee percentage");
console.log("    {");
console.log("      from: eth.accounts[0],");
console.log("      data: exampleDAppBytecode,");
console.log("      gas: 4000000");
console.log("    },");
console.log("    function(e, contract) {");
console.log("      if (e) {");
console.log("        console.log(\"Error:\", e);");
console.log("      } else if (contract.address) {");
console.log("        console.log(\"Contract deployed at:\", contract.address);");
console.log("      }");
console.log("    }");
console.log("  );");
console.log("");

console.log("==============================================");
console.log("STEP 3: Verify Deployment");
console.log("==============================================");
console.log("");
console.log("After deployment completes (wait for mining):");
console.log("");
console.log("  // Check transaction receipt");
console.log("  eth.getTransactionReceipt(exampleDApp.transactionHash);");
console.log("");
console.log("  // Get contract instance");
console.log("  var dapp = ExampleDAppContract.at(exampleDApp.address);");
console.log("");
console.log("  // Verify fee configuration");
console.log("  var config = dapp.getFeeConfig();");
console.log("  console.log(\"Enabled:\", config[0]);");
console.log("  console.log(\"Recipient:\", config[1]);");
console.log("  console.log(\"Percent:\", config[2].toString());");
console.log("");

console.log("==============================================");
console.log("STEP 4: Test the Contract");
console.log("==============================================");
console.log("");
console.log("Test that fee sharing works:");
console.log("");
console.log("  // Record initial balance");
console.log("  var balanceBefore = eth.getBalance(\"" + recipientAddress + "\");");
console.log("  console.log(\"Balance before:\", web3.fromWei(balanceBefore, \"ether\"));");
console.log("");
console.log("  // Execute some transactions");
console.log("  dapp.updateScore(100, { from: eth.accounts[0], gas: 100000 });");
console.log("  dapp.incrementScore({ from: eth.accounts[0], gas: 100000 });");
console.log("  dapp.updateScore(200, { from: eth.accounts[0], gas: 100000 });");
console.log("");
console.log("  // Wait for blocks to be mined");
console.log("  admin.sleepBlocks(2);");
console.log("");
console.log("  // Check balance after");
console.log("  var balanceAfter = eth.getBalance(\"" + recipientAddress + "\");");
console.log("  console.log(\"Balance after:\", web3.fromWei(balanceAfter, \"ether\"));");
console.log("");
console.log("  // Calculate fees received");
console.log("  var feesReceived = balanceAfter - balanceBefore;");
console.log("  console.log(\"Fees received:\", web3.fromWei(feesReceived, \"ether\"), \"HALO\");");
console.log("");

console.log("==============================================");
console.log("Additional Commands");
console.log("==============================================");
console.log("");
console.log("Update fee recipient:");
console.log("  dapp.updateRecipient(\"0xNewAddress\", { from: eth.accounts[0], gas: 100000 });");
console.log("");
console.log("Update fee percentage:");
console.log("  dapp.updatePercent(75, { from: eth.accounts[0], gas: 100000 });");
console.log("");
console.log("Disable fee sharing:");
console.log("  dapp.disableFeeSharing({ from: eth.accounts[0], gas: 100000 });");
console.log("");
console.log("Check user score:");
console.log("  dapp.getScore(eth.accounts[0]);");
console.log("");
console.log("Check total interactions:");
console.log("  dapp.totalInteractions();");
console.log("");

console.log("==============================================");
console.log("Quick Reference");
console.log("==============================================");
console.log("");
console.log("Contract Files:");
console.log("  contracts/HaloFeeSharing.sol    - Base contract");
console.log("  contracts/ExampleDApp.sol       - Example implementation");
console.log("  contracts/DEPLOYMENT_GUIDE.md   - Complete guide");
console.log("");
console.log("Fee Distribution on Halo Chain:");
console.log("  40% - Burned (deflationary)");
console.log("  30% - Miner (NEVER reduced by contract sharing)");
console.log("  20% - Ecosystem Fund (may be shared with contracts)");
console.log("  10% - Reserve Fund (NEVER reduced)");
console.log("");
console.log("With 50% contract fee sharing:");
console.log("  40% - Burned");
console.log("  30% - Miner (unchanged - keeps network secure)");
console.log("  10% - Contract (from ecosystem's 20%)");
console.log("  10% - Ecosystem Fund (reduced from 20%)");
console.log("  10% - Reserve Fund");
console.log("");

console.log("==============================================");
console.log("For detailed instructions, see:");
console.log("  contracts/DEPLOYMENT_GUIDE.md");
console.log("==============================================");
console.log("");

// Export helper functions
function deployExampleDApp(bytecode, recipient, percent) {
    if (!bytecode || bytecode.length < 10) {
        console.log("Error: Invalid bytecode. Please compile contracts first.");
        return;
    }

    if (!recipient || recipient == "0x0000000000000000000000000000000000000000") {
        console.log("Error: Invalid recipient address.");
        return;
    }

    if (percent < 0 || percent > 100) {
        console.log("Error: Percent must be between 0 and 100.");
        return;
    }

    console.log("Deploying ExampleDApp...");
    console.log("  Recipient:", recipient);
    console.log("  Percent:", percent + "%");

    var ExampleDAppContract = eth.contract(exampleDAppAbi);

    var contract = ExampleDAppContract.new(
        recipient,
        percent,
        {
            from: eth.accounts[0],
            data: bytecode,
            gas: 4000000
        },
        function(e, contract) {
            if (e) {
                console.log("Deployment error:", e);
            } else if (contract.address) {
                console.log("ExampleDApp deployed successfully!");
                console.log("  Address:", contract.address);
                console.log("  Transaction:", contract.transactionHash);
                console.log("");
                console.log("Verify deployment:");
                console.log("  eth.getTransactionReceipt(\"" + contract.transactionHash + "\")");
            }
        }
    );

    return contract;
}

console.log("Helper function loaded:");
console.log("  deployExampleDApp(bytecode, recipientAddress, feePercent)");
console.log("");
console.log("Example usage:");
console.log("  var bytecode = \"0x...\"; // From compilation");
console.log("  var contract = deployExampleDApp(bytecode, eth.accounts[0], 50);");
console.log("");
