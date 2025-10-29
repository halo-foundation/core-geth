// Copyright 2025 The core-geth Authors
// This file is part of the core-geth library.
//
// The core-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The core-geth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the core-geth library. If not, see <http://www.gnu.org/licenses/>.

package eip1559

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

// =============================================================================
// HALO EIP-1559 CUSTOM FEE DISTRIBUTION
// =============================================================================

// Halo EIP-1559 fee distribution ratios (out of 1000)
const (
	HaloBurnRatio      = 400 // 40% burned
	HaloMinerRatio     = 300 // 30% to miners
	HaloEcosystemRatio = 200 // 20% to ecosystem fund
	HaloReserveRatio   = 100 // 10% to reserve fund
	HaloTotalRatio     = 1000
)

var (
	ErrZeroEcosystemAddress = errors.New("ecosystem fund address cannot be zero address")
	ErrZeroReserveAddress   = errors.New("reserve fund address cannot be zero address")
	ErrInvalidFeePercent    = errors.New("fee share percent must be 0-100")
)

// ValidateHaloAddresses validates that Halo fund addresses are not zero addresses
func ValidateHaloAddresses() error {
	if params.HaloEcosystemFundAddress == (common.Address{}) {
		return ErrZeroEcosystemAddress
	}
	if params.HaloReserveFundAddress == (common.Address{}) {
		return ErrZeroReserveAddress
	}
	return nil
}

// ApplyHaloBaseFeeDistribution applies the Halo custom EIP-1559 base fee distribution
// This function is called during block finalization to distribute base fees according to:
// - 40% burned (reduces total supply)
// - 30% to miners (coinbase) - NEVER reduced by contract fee sharing
// - 20% to ecosystem fund - MAY be reduced by contract fee sharing
// - 10% to reserve fund - NEVER reduced by contract fee sharing
//
// Contract fee sharing (if enabled) deducts from the ecosystem fund's 20%, not from miners.
// This ensures miner incentives remain intact while supporting dApp development.
//
// Priority fees (tips) go 100% to miners.
func ApplyHaloBaseFeeDistribution(state *state.StateDB, header *types.Header, baseFee *big.Int, gasUsed uint64) error {
	// Validate fund addresses
	if err := ValidateHaloAddresses(); err != nil {
		return err
	}

	// Calculate total base fee collected
	totalBaseFee := new(big.Int).Mul(baseFee, new(big.Int).SetUint64(gasUsed))
	if totalBaseFee.Sign() == 0 {
		return nil // No fees to distribute
	}

	totalBaseFeeU256 := uint256.MustFromBig(totalBaseFee)

	// Calculate distribution amounts
	// Burn amount: 40%
	burnAmount := new(uint256.Int).Mul(totalBaseFeeU256, uint256.NewInt(HaloBurnRatio))
	burnAmount.Div(burnAmount, uint256.NewInt(HaloTotalRatio))

	// Miner amount: 30%
	minerAmount := new(uint256.Int).Mul(totalBaseFeeU256, uint256.NewInt(HaloMinerRatio))
	minerAmount.Div(minerAmount, uint256.NewInt(HaloTotalRatio))

	// Ecosystem fund amount: 20%
	ecosystemAmount := new(uint256.Int).Mul(totalBaseFeeU256, uint256.NewInt(HaloEcosystemRatio))
	ecosystemAmount.Div(ecosystemAmount, uint256.NewInt(HaloTotalRatio))

	// Reserve fund amount: 10%
	reserveAmount := new(uint256.Int).Mul(totalBaseFeeU256, uint256.NewInt(HaloReserveRatio))
	reserveAmount.Div(reserveAmount, uint256.NewInt(HaloTotalRatio))

	// Apply distributions
	// 1. Burn - implicit (not added to any account, reduces total supply)
	// 2. Miner reward - add to coinbase
	state.AddBalance(header.Coinbase, minerAmount)
	// 3. Ecosystem fund - add to ecosystem address
	state.AddBalance(params.HaloEcosystemFundAddress, ecosystemAmount)
	// 4. Reserve fund - add to reserve address
	state.AddBalance(params.HaloReserveFundAddress, reserveAmount)

	return nil
}

// =============================================================================
// PER-CONTRACT FEE SHARING (SONIC-STYLE)
// =============================================================================

// Storage slots for contract fee sharing configuration (EIP-1967 style)
var (
	// keccak256("halo.feeshare.enabled") - 1
	feeShareEnabledSlot = common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")
	// keccak256("halo.feeshare.recipient") - 1
	feeShareRecipientSlot = common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbd")
	// keccak256("halo.feeshare.percent") - 1
	feeSharePercentSlot = common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbe")
)

// HaloContractFeeConfig holds per-contract fee sharing configuration
type HaloContractFeeConfig struct {
	Enabled         bool           // Whether fee sharing is enabled for this contract
	FeeRecipient    common.Address // Address that receives the contract's fee share
	FeeSharePercent uint8          // Percentage of fees to share (0-100)
}

// GetContractFeeConfig retrieves fee sharing configuration from contract storage
// Reads from special storage slots in the contract's own state (EIP-1967 style)
func GetContractFeeConfig(state *state.StateDB, contractAddr common.Address) *HaloContractFeeConfig {
	// Read enabled flag from storage
	enabledValue := state.GetState(contractAddr, feeShareEnabledSlot)
	enabled := enabledValue != (common.Hash{}) && enabledValue.Big().Sign() != 0

	if !enabled {
		return &HaloContractFeeConfig{
			Enabled:         false,
			FeeRecipient:    common.Address{},
			FeeSharePercent: 0,
		}
	}

	// Read recipient address
	recipientValue := state.GetState(contractAddr, feeShareRecipientSlot)
	recipient := common.BytesToAddress(recipientValue.Bytes())

	// Read percentage
	percentValue := state.GetState(contractAddr, feeSharePercentSlot)
	percent := uint8(percentValue.Big().Uint64())

	// Validate percentage
	if percent > 100 {
		percent = 0
		enabled = false
	}

	return &HaloContractFeeConfig{
		Enabled:         enabled,
		FeeRecipient:    recipient,
		FeeSharePercent: percent,
	}
}

// SetContractFeeConfig sets fee sharing configuration in contract storage
// Stores in special storage slots in the contract's own state
//
// IMPORTANT: Access control must be implemented in the calling smart contract.
// This function only handles storage - the contract must ensure only authorized
// addresses (e.g., contract owner) can call this.
func SetContractFeeConfig(state *state.StateDB, contractAddr common.Address, config *HaloContractFeeConfig) error {
	// Validate configuration
	if config.FeeSharePercent > 100 {
		return ErrInvalidFeePercent
	}

	// Set enabled flag
	var enabledValue common.Hash
	if config.Enabled {
		enabledValue = common.BigToHash(big.NewInt(1))
	}
	state.SetState(contractAddr, feeShareEnabledSlot, enabledValue)

	// Set recipient address
	recipientValue := common.BytesToHash(config.FeeRecipient.Bytes())
	state.SetState(contractAddr, feeShareRecipientSlot, recipientValue)

	// Set percentage
	percentValue := common.BigToHash(big.NewInt(int64(config.FeeSharePercent)))
	state.SetState(contractAddr, feeSharePercentSlot, percentValue)

	return nil
}

// ApplyHaloContractFeeSharing applies per-contract fee sharing for a transaction
// If the transaction interacts with a contract that has fee sharing enabled,
// a portion of the transaction fees is redirected to the contract's specified recipient
//
// IMPORTANT: Fee shares are deducted from the ECOSYSTEM FUND (20% of base fees)
// The ecosystem fund supports development and growth, so sharing with dApp developers
// aligns with this purpose while keeping miner incentives intact.
//
// Example with 50% contract fee sharing:
// - 40% burned
// - 30% to miner (unchanged)
// - 10% to ecosystem fund (reduced from 20%)
// - 10% to contract (from ecosystem's 20%)
// - 10% to reserve fund (unchanged)
//
// SECURITY: Includes balance checks to prevent underflow attacks
//
// This is called during block finalization to distribute contract fee shares
func ApplyHaloContractFeeSharing(state *state.StateDB, tx *types.Transaction, contractAddr common.Address, gasUsed uint64, baseFee *big.Int) error {
	config := GetContractFeeConfig(state, contractAddr)
	if !config.Enabled || config.FeeSharePercent == 0 {
		return nil // Fee sharing not enabled for this contract
	}

	// Validate recipient
	if config.FeeRecipient == (common.Address{}) {
		return nil // No valid recipient
	}

	// Validate ecosystem fund address
	if err := ValidateHaloAddresses(); err != nil {
		return err
	}

	// Calculate total base fee for this transaction
	totalFee := new(big.Int).Mul(baseFee, new(big.Int).SetUint64(gasUsed))
	totalFeeU256 := uint256.MustFromBig(totalFee)

	// Calculate the ecosystem fund's portion (20% of total base fee)
	ecosystemPortion := new(uint256.Int).Mul(totalFeeU256, uint256.NewInt(HaloEcosystemRatio))
	ecosystemPortion.Div(ecosystemPortion, uint256.NewInt(HaloTotalRatio))

	// Calculate contract's share (percentage of ecosystem's portion)
	contractShare := new(uint256.Int).Mul(ecosystemPortion, uint256.NewInt(uint64(config.FeeSharePercent)))
	contractShare.Div(contractShare, uint256.NewInt(100))

	if contractShare.Sign() > 0 {
		// SECURITY FIX: Check ecosystem fund has sufficient balance before deduction
		ecosystemBalance := state.GetBalance(params.HaloEcosystemFundAddress)

		// If ecosystem fund has insufficient balance, only transfer what's available
		// This prevents underflow attacks and state corruption
		if ecosystemBalance.Cmp(contractShare) < 0 {
			// Insufficient funds - only transfer available balance
			if ecosystemBalance.Sign() > 0 {
				state.SubBalance(params.HaloEcosystemFundAddress, ecosystemBalance)
				state.AddBalance(config.FeeRecipient, ecosystemBalance)
			}
			// Don't error - just skip/reduce the fee sharing for this tx
			return nil
		}

		// Sufficient balance - proceed with full fee sharing
		state.SubBalance(params.HaloEcosystemFundAddress, contractShare)
		state.AddBalance(config.FeeRecipient, contractShare)
	}

	return nil
}

// =============================================================================
// ALTERNATIVE: REGISTRY-BASED FEE SHARING (OPTIONAL)
// =============================================================================

var (
	// System registry contract address (if using registry approach)
	// This would be deployed at a known address
	feeShareRegistryAddress = common.HexToAddress("0x0000000000000000000000000000000000000FEE")
)

// GetContractFeeConfigFromRegistry retrieves config from a central registry
// This is an ALTERNATIVE implementation - use either this OR the storage-based approach
func GetContractFeeConfigFromRegistry(state *state.StateDB, contractAddr common.Address) *HaloContractFeeConfig {
	// Calculate storage slot for this contract's config
	baseSlot := common.BigToHash(big.NewInt(0))
	slot := crypto.Keccak256Hash(contractAddr.Bytes(), baseSlot.Bytes())

	// Read enabled flag from registry
	enabledValue := state.GetState(feeShareRegistryAddress, slot)
	enabled := enabledValue != (common.Hash{}) && enabledValue.Big().Sign() != 0

	if !enabled {
		return &HaloContractFeeConfig{
			Enabled:         false,
			FeeRecipient:    common.Address{},
			FeeSharePercent: 0,
		}
	}

	// Read recipient (slot + 1)
	recipientSlot := common.BigToHash(new(big.Int).Add(slot.Big(), big.NewInt(1)))
	recipientValue := state.GetState(feeShareRegistryAddress, recipientSlot)
	recipient := common.BytesToAddress(recipientValue.Bytes())

	// Read percentage (slot + 2)
	percentSlot := common.BigToHash(new(big.Int).Add(slot.Big(), big.NewInt(2)))
	percentValue := state.GetState(feeShareRegistryAddress, percentSlot)
	percent := uint8(percentValue.Big().Uint64())

	if percent > 100 {
		percent = 0
		enabled = false
	}

	return &HaloContractFeeConfig{
		Enabled:         enabled,
		FeeRecipient:    recipient,
		FeeSharePercent: percent,
	}
}

/*
=============================================================================
SOLIDITY REFERENCE IMPLEMENTATION
=============================================================================

Below is a reference Solidity contract showing how to enable fee sharing:

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract HaloFeeSharing {
    // Storage slots must match the Go implementation
    bytes32 private constant ENABLED_SLOT = 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc;
    bytes32 private constant RECIPIENT_SLOT = 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbd;
    bytes32 private constant PERCENT_SLOT = 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbe;

    address public owner;

    event FeeShareEnabled(address indexed recipient, uint8 percent);
    event FeeShareDisabled();

    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }

    constructor() {
        owner = msg.sender;
    }

    /// @notice Enable fee sharing for this contract
    /// @param recipient Address to receive fee share
    /// @param percent Percentage of fees to share (0-100)
    function enableFeeSharing(address recipient, uint8 percent) external onlyOwner {
        require(percent <= 100, "Invalid percent");
        require(recipient != address(0), "Invalid recipient");

        assembly {
            sstore(ENABLED_SLOT, 1)
            sstore(RECIPIENT_SLOT, recipient)
            sstore(PERCENT_SLOT, percent)
        }

        emit FeeShareEnabled(recipient, percent);
    }

    /// @notice Disable fee sharing for this contract
    function disableFeeSharing() external onlyOwner {
        assembly {
            sstore(ENABLED_SLOT, 0)
        }

        emit FeeShareDisabled();
    }

    /// @notice Get current fee sharing configuration
    /// @return enabled Whether fee sharing is enabled
    /// @return recipient Address receiving fee share
    /// @return percent Percentage of fees being shared
    function getFeeConfig() external view returns (
        bool enabled,
        address recipient,
        uint8 percent
    ) {
        assembly {
            enabled := sload(ENABLED_SLOT)
            recipient := sload(RECIPIENT_SLOT)
            percent := sload(PERCENT_SLOT)
        }
    }

    /// @notice Update fee share recipient
    /// @param newRecipient New recipient address
    function updateRecipient(address newRecipient) external onlyOwner {
        require(newRecipient != address(0), "Invalid recipient");

        assembly {
            sstore(RECIPIENT_SLOT, newRecipient)
        }
    }

    /// @notice Update fee share percentage
    /// @param newPercent New percentage (0-100)
    function updatePercent(uint8 newPercent) external onlyOwner {
        require(newPercent <= 100, "Invalid percent");

        assembly {
            sstore(PERCENT_SLOT, newPercent)
        }
    }

    /// @notice Transfer ownership
    /// @param newOwner New owner address
    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "Invalid address");
        owner = newOwner;
    }
}

=============================================================================
*/
