package filters

import (
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
)

type FilterQuery struct {
	BlockHash *crypto.Hash           // used by eth_getLogs, return logs only from block with this hash
	FromBlock *big.Int               // beginning of the queried range, nil means genesis block
	ToBlock   *big.Int               // end of the range, nil means latest block
	Addresses []crypto.CommonAddress // restricts matches to events created by specific contracts

	// The Topic list restricts matches to particular event topics. Each event has a list
	// of topics. Topics matches a prefix of that list. An empty element slice matches any
	// topic. Non-empty elements represent an alternative that matches any of the
	// contained topics.
	//
	// Examples:
	// {} or nil          matches any topic list
	// {{A}}              matches topic A in first position
	// {{}, {B}}          matches any topic in first position AND B in second position
	// {{A}, {B}}         matches topic A in first position AND B in second position
	// {{A, B}, {C, D}}   matches topic (A OR B) in first position AND (C OR D) in second position
	Topics [][]crypto.Hash
}
