// Copyright 2016 The go-ethereum Authors
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

package vm

import (
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/types"
	"math/big"
	"testing"
)

type dummyStatedb struct {
	State
}

func (*dummyStatedb) GetRefund() uint64 { return 1337 }

func TestStoreCapture(t *testing.T) {
	var (
		env      = NewEVM(Context{}, &dummyStatedb{}, &VMConfig{})
		logger   = NewStructLogger(nil)
		mem      = NewMemory()
		stack    = newstack()
		contract = NewContract(crypto.CommonAddress{}, types.ChainIdType(0), 0, new(big.Int), nil)
	)
	stack.push(big.NewInt(1))
	stack.push(big.NewInt(0))
	var index crypto.Hash
	logger.CaptureState(env, 0, SSTORE, 0, 0, mem, stack, contract, 0, nil)
	if len(logger.changedValues[contract.ContractAddr]) == 0 {
		t.Fatalf("expected exactly 1 changed value on address %x, got %d", contract.ContractAddr, len(logger.changedValues[contract.ContractAddr]))
	}
	exp := crypto.BigToHash(big.NewInt(1))
	if logger.changedValues[contract.ContractAddr][index] != exp {
		t.Errorf("expected %x, got %x", exp, logger.changedValues[contract.ContractAddr][index])
	}
}
