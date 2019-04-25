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

package evm

import (
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
	"math/big"


)

// ChainContext supports retrieving headers and consensus parameters from the
// current blockchain to be used during transaction processing.
type ChainContext interface {
	// GetHeader returns the hash corresponding to their hash.
	GetHeader(crypto.Hash, uint64) *types.BlockHeader
}

// NewEVMContext creates a new context for use in the EVM.
func NewEVMContext(msg *types.Transaction, header *types.BlockHeader, sender *crypto.CommonAddress, chain ChainContext) vm.Context {
	return vm.Context{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		Origin:      *sender,
		BlockNumber: big.NewInt(int64(header.Height)),
		Time:        big.NewInt(int64(header.Timestamp)),
		GasLimit:    header.GasLimit.Uint64(),
		GasPrice:    new(big.Int).Set(msg.GasPrice()),
	}
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(ref *types.BlockHeader, chain ChainContext) func(n uint64) crypto.Hash {
	var cache map[uint64]crypto.Hash
	return func(n uint64) crypto.Hash {
		// If there's no hash cache yet, make one
		if cache == nil {
			cache = map[uint64]crypto.Hash{
				uint64(ref.Height - 1): ref.PreviousHash,
			}
		}
		// Try to fulfill the request from the cache
		if hash, ok := cache[n]; ok {
			return hash
		}
		// Not cached, iterate the blocks and cache the hashes
		for header := chain.GetHeader(ref.PreviousHash, uint64(ref.Height - 1)); header != nil; header = chain.GetHeader(header.PreviousHash, uint64(header.Height - 1)) {
			cache[uint64(header.Height - 1)] = header.PreviousHash
			if n == uint64(header.Height - 1) {
				return header.PreviousHash
			}
		}
		return crypto.Hash{}
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.State, addr crypto.CommonAddress, amount *big.Int) bool {
	return db.GetBalance(&addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.State, sender, to crypto.CommonAddress, amount *big.Int) {
	db.SubBalance(&sender, amount)
	db.AddBalance(&to, amount)
}
