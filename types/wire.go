package types

import "github.com/drep-project/DREP-Chain/crypto"

//本模块的消息只能在调用本模块（chain及对应的子模块）的函数中使用
const (
	MsgTypeBlockReq     = 1 //同步块请求
	MsgTypeBlockResp    = 2 //同步块回复
	MsgTypeBlock        = 3 //新块通知
	MsgTypeTransaction  = 4 //广播交易
	MsgTypePeerState    = 5 //Peer状态回复/或者状态通知
	MsgTypePeerStateReq = 6 //peer状态请求
	MsgTypeHeaderReq    = 7 //请求区块头
	MsgTypeHeaderRsp    = 8 //请求区块头回复

	MaxMsgSize = 20 << 20 //每个消息最大大小20MB
)

var NumberOfMsg = 9 //本模块定义的消息个数

type Transactions []Transaction

type HeaderReq struct {
	FromHeight uint64
	ToHeight   uint64
}

type HeaderRsp struct {
	//Heights []uint64
	Headers []BlockHeader
}

type BlockReq struct {
	BlockHashs []crypto.Hash
}

type BlockResp struct {
	Height uint64
	Blocks []*Block
}

type PeerState struct {
	Height uint64
}

type PeerStateReq struct {
	Height uint64
}

//type Inv struct {
//	Hashes []crypto.Hash
//}
//
//func (inv *Inv) AddInv(hash *crypto.Hash) {
//	inv.Hashes = append(inv.Hashes, *hash)
//}
