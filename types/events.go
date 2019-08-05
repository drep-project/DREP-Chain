package types


import (
	"github.com/drep-project/drep-chain/crypto"
)

// NewTxsEvent is posted when a batch of transactions enter the transaction pool.
type NewTxsEvent struct{ Txs []*Transaction }

// PendingLogsEvent is posted pre mining and notifies of pending logs.
type PendingLogsEvent struct {
	Logs []*Log
}

// NewMinedBlockEvent is posted when a block has been imported.
type NewMinedBlockEvent struct{ Block *Block }

// RemovedLogsEvent is posted when a reorg happens
type RemovedLogsEvent struct{ Logs []*Log }

type ChainEvent struct {
	Block *Block
	Hash  crypto.Hash
	Logs  []*Log
}

type ChainSideEvent struct {
	Block *Block
}

type ChainHeadEvent struct{ Block *Block }


