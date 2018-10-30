package bean

import "BlockChainTest/mycrypto"

type PeerInfo struct {
	Pk                   *mycrypto.Point
	Ip                   string
	Port                 int32
}

type PeerInfoList struct {
	List                 []*PeerInfo
}

type BlockReq struct {
	Pk                   *mycrypto.Point
	Height               int64
}

type BlockResp struct {
	Height               int64
	Blocks               []*Block
}

type Ping struct {
	Pk                   *mycrypto.Point
}

type Pong struct {
	Pk                   *mycrypto.Point
}

type OfflinePeers struct {
	List                 []*PeerInfo
}