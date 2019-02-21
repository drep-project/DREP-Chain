package types

import (
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)
const (
	MsgTypeBlockReq  = 3
	MsgTypeBlockResp = 4
	MsgTypeBlock = 5
	MsgTypeTransaction = 6
	MsgTypePeerState = 7
	MsgTypeReqPeerState = 8
)
type BlockReq struct {
	Pk                   *secp256k1.PublicKey
	Height               int64
}

type BlockResp struct {
	Height               int64
	Blocks               []*Block
}

type ReqPeerState struct {
	Height int64
}

type PeerState struct {
	Height int64
}
