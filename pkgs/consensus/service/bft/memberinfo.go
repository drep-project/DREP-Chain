package bft

import (
	"github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	chainType "github.com/drep-project/DREP-Chain/types"
)

type MemberInfo struct {
	Peer     types.IPeerInfo
	Producer *chainType.Producer
	Status   int
	IsMe     bool
	IsLeader bool
	IsOnline bool
}
