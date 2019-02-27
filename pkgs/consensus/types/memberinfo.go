package types

import (
	p2pTypes "github.com/drep-project/drep-chain/network/types"
)

type MemberInfo struct {
	Peer     *p2pTypes.Peer
	Producer *Producer
}
