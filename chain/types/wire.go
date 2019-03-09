package types

import "github.com/drep-project/drep-chain/crypto"

const (
	MsgTypeBlockReq  = 3
	MsgTypeBlockResp = 4
	MsgTypeBlock = 5
	MsgTypeTransaction = 6
	MsgTypePeerState = 7
	MsgTypeReqPeerState = 8
	MsgTypeInv = 8
)
type BlockReq struct {
	StartHash			crypto.Hash  //may not exist
	StopHash			crypto.Hash
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

type Inv struct {
	Hashes []crypto.Hash
}

func (inv *Inv) AddInv(hash *crypto.Hash){
	inv.Hashes = append(inv.Hashes, *hash)
}
