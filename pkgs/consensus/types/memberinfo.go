package types

import (
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
)

const (
	OnLine = iota
	OffLine = iota
)
type MemberInfo struct {
	Peer     	*p2pTypes.Peer
	Producer 	*chainTypes.Producer
	Status 		int
	IsMe		bool
	IsLeader 	bool
	IsOnline	bool
}
