// Copyright 2014 The go-ethereum Authors
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
package service

import (
	"errors"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/pkgs/evm"
	"math/big"

	"github.com/drep-project/drep-chain/pkgs/evm/vm"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

/*
The State Transitioning Model

A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all the necessary work to work out a valid new state root.

1) Nonce handling
2) Pre pay gas
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==
  4a) Attempt to run transaction data
  4b) If valid, use result as code for the new state object
== end ==
5) Run Script section
6) Derive new state root
*/
type StateTransition struct {
	gp         *GasPool
	tx        *types.Transaction
	gas        uint64
	gasPrice   *big.Int
	initialGas uint64
	value      *big.Int
	data       []byte
	state      *vm.State
	header *types.BlockHeader
	bc evm.ChainContext
	vmService       evm.Vm
	databaseService *database.DatabaseService
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(databaseService *database.DatabaseService, vmService evm.Vm,tx *types.Transaction, state *vm.State, header *types.BlockHeader, bc evm.ChainContext, gp *GasPool) *StateTransition {
	return &StateTransition{
		gp:       gp,
		tx:       tx,
		gasPrice: tx.GasPrice(),
		value:    &tx.Data.Amount,
		data:     tx.Data.Data,
		state:    state,
		header:	  header,
		bc:		  bc,
		vmService: vmService,
		databaseService:databaseService,
	}
}

// to returns the recipient of the message.
func (st *StateTransition) to() crypto.CommonAddress {
	if st.tx == nil || st.tx.To() == nil || st.tx.To().IsEmpty() /* contract creation */ {
		return crypto.CommonAddress{}
	}
	return *st.tx.To()
}

func (st *StateTransition) useGas(amount uint64) error {
	if st.gas < amount {
		return vm.ErrOutOfGas
	}
	st.gas -= amount

	return nil
}

func (st *StateTransition) buyGas() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.tx.Gas()), st.gasPrice)
	if st.state.GetBalance(st.tx.From()).Cmp(mgval) < 0 {
		return errInsufficientBalanceForGas
	}
	if err := st.gp.SubGas(st.tx.Gas()); err != nil {
		return err
	}
	st.gas += st.tx.Gas()

	st.initialGas = st.tx.Gas()
	st.state.SubBalance(st.tx.From(), mgval)
	return nil
}

func (st *StateTransition) preCheck() error {
	// Make sure this transaction's nonce is correct.
	nonce := st.state.GetNonce(st.tx.From())
	if nonce < st.tx.Nonce() {
		return  errors.New("nonce too high")
	} else if nonce > st.tx.Nonce() {
		return  errors.New("nonce too low")
	}
	return st.buyGas()
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the used gas. It returns an error if failed.
// An error indicates a consensus issue.
func (st *StateTransition) TransitionVmTxDb() (ret []byte, failed bool, err error) {
	ret, fail, err := st.vmService.Eval(st.state, st.tx, st.header, st.bc, st.gas, st.value)
	return ret, fail, err
}

func (st *StateTransition) TransitionTransferDb() (ret []byte, failed bool, err error) {
	from := st.tx.From()
	originBalance := st.databaseService.GetBalance(from, true)
	toBalance :=  st.databaseService.GetBalance(st.tx.To(), true)
	leftBalance := originBalance.Sub(originBalance, st.tx.Amount())
	if leftBalance.Sign() <0 {
		return nil,  false, errors.New("from addr in tx is not eenough")
	}
	addBalance := toBalance.Add(toBalance, st.tx.Amount())
	st.databaseService.PutBalance(from, leftBalance, true)
	st.databaseService.PutBalance(st.tx.To(), addBalance, true)
	st.databaseService.PutNonce(from, st.tx.Nonce()+1, true)
	return nil,  true, nil
}

func (st *StateTransition) TransitionAliasDb() (ret []byte, failed bool, err error) {
	from := st.tx.From()
	alias := st.tx.GetData()
	err = st.databaseService.AliasSet(from,string(alias))
	return nil,  true, err
}

func (st *StateTransition) refundGas() {
	// Apply refund counter, capped to half of the used gas.
	refund := st.gasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return DREP for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(st.tx.From(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}
