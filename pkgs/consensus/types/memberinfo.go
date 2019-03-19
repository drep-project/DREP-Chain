package types

import (
	"github.com/drep-project/drep-chain/chain/types"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
)

const (
	OnLine = iota
	OffLine = iota
)
type MemberInfo struct {
	Peer     	*p2pTypes.Peer
	Producer 	*types.Producer
	Status 		int
	IsMe		bool
	IsLeader 	bool
	IsOnline	bool
}
