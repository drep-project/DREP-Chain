package types

import "github.com/drep-project/drep-chain/crypto"

const (
	MsgTypeBlockReq     = 3
	MsgTypeBlockResp    = 4
	MsgTypeBlock        = 5
	MsgTypeTransaction  = 6
	MsgTypePeerState    = 7
	MsgTypeReqPeerState = 8
	//MsgTypeInv = 8
	MsgTypeHeaderReq = 9
	MsgTypeHeaderRsp = 10
)

type HeaderReq struct {
	FromHeight int64
	ToHeight   int64
}

type HeaderRsp struct {
	//Heights []int64
	Headers   []BlockHeader
}

type BlockReq struct {
	BlockHashs []crypto.Hash
}

type BlockResp struct {
	Height int64
	Blocks []*Block
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

func (inv *Inv) AddInv(hash *crypto.Hash) {
	inv.Hashes = append(inv.Hashes, *hash)
}
