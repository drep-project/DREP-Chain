package types

import (
	p2pTypes "github.com/drep-project/drep-chain/network/types"
)
type Member struct {
	Peer *p2pTypes.Peer
	Produce *Produce
}
