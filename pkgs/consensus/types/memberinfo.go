package types

import "github.com/drep-project/drep-chain/crypto/secp256k1"

const (
	OnLine = iota
	OffLine = iota
)
type MemberInfo struct {
	Peer     *PeerInfo
	Producer *Producer
	Status   int
	IsMe     bool
	IsLeader bool
	IsOnline bool
}

type Producer struct {
	Public *secp256k1.PublicKey
	Ip     string
}
