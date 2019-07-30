package types

import (
	"github.com/drep-project/drep-chain/network/p2p"
	"github.com/drep-project/drep-chain/types"
)

type IConsensusEngin interface {
	Run() (*types.Block, error)
	ReceiveMsg(peer *PeerInfo, rw p2p.MsgReadWriter) error
}
