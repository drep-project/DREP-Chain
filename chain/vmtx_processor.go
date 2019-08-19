// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package chain

import (
	"fmt"
	"math/big"

	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/pkgs/evm"
	"github.com/drep-project/drep-chain/types"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	chainService *ChainService
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(chainservice *ChainService) *StateProcessor {
	return &StateProcessor{
		chainService: chainservice,
	}
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func (stateProcessor *StateProcessor) ApplyTransaction(db *database.Database, bc evm.ChainContext, gp *GasPool, header *types.BlockHeader, tx *types.Transaction, from *crypto.CommonAddress, usedGas *uint64) (*types.Receipt, uint64, error) {
	// Apply the transaction to the current state (included in the env)
	_, gas, _, failed, err := stateProcessor.ApplyMessage(db, tx, from, header, bc, gp)
	if err != nil {
		return nil, 0, err
	}
	*usedGas += gas

	root := db.GetStateRoot()
	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = *tx.TxHash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if tx.To() == nil || tx.To().IsEmpty() {
		receipt.ContractAddress = crypto.CreateAddress(*from, tx.Nonce())
		fmt.Println(receipt.ContractAddress)
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = db.GetLogs(*tx.TxHash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	//receipt.BlockHash = *header.Hash()
	receipt.BlockNumber = header.Height
	return receipt, gas, err
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func (stateProcessor *StateProcessor) ApplyMessage(db *database.Database, tx *types.Transaction, from *crypto.CommonAddress, header *types.BlockHeader, bc evm.ChainContext, gp *GasPool) ([]byte, uint64, uint64, bool, error) {
	stateTransaction := NewStateTransition(db, stateProcessor.chainService.VmService, tx, from, header, bc, gp)
	if err := stateTransaction.preCheck(); err != nil {
		return nil, 0, 0, false, err
	}

	// Pay intrinsic gastx
	gas, err := tx.IntrinsicGas()
	if err != nil {
		return nil, 0, 0, false, err
	}

	if err = stateTransaction.useGas(gas); err != nil {
		return nil, 0, 0, false, err
	}

	var ret []byte
	var fail bool
	if tx.Type() == types.TransferType {
		ret, fail, err = stateTransaction.TransitionTransferDb()
	} else if tx.Type() == types.CallContractType || tx.Type() == types.CreateContractType {
		ret, fail, err = stateTransaction.TransitionVmTxDb()
	} else if tx.Type() == types.SetAliasType {
		ret, fail, err = stateTransaction.TransitionAliasDb()
	} else {
		return nil, 0, 0, false, ErrUnsupportTxType
	}

	stateTransaction.refundGas()
	gasFee := new(big.Int).Mul(new(big.Int).SetUint64(stateTransaction.gasUsed()), stateTransaction.gasPrice).Uint64()
	return ret, stateTransaction.gasUsed(), gasFee, fail, err
}
