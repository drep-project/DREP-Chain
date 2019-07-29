package types

import (
	"github.com/drep-project/drep-chain/crypto"
)

type Log struct {
	// Consensus fields:
	// address of the contract that generated the event
	Address crypto.CommonAddress
	// list of topics provided by the contract.
	Topics []crypto.Hash
	// supplied by the contract, usually ABI-encoded
	Data []byte

	// Derived fields. These fields are filled in by the node but not secured by consensus.
	ChainId ChainIdType
	// hash of the transaction
	TxHash crypto.Hash
	// block in which the transaction was included
	Height uint64
	// index of the transaction in the block
	TxIndex uint

	// The Removed field is true if this log was reverted due to a chain reorganisation.
	// You must pay attention to this field if you receive logs through a filter query.
	Removed bool
}
