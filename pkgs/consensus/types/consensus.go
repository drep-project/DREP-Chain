package types

import (
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/types"
)

type IConsensusEngine interface {
	Run(key *secp256k1.PrivateKey) (*types.Block, error)
	ReceiveMsg(peer *PeerInfo, t uint64, buf []byte)
}
