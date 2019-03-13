package types

import "github.com/drep-project/drep-chain/network/types"

const (
	MsgTypeSetUp  = 11
	MsgTypeCommitment = 12
	MsgTypeResponse = 13
	MsgTypeChallenge = 14
	MsgTypeFail = 15
)

type RouteMsgWrap struct {
	Peer *types.Peer
	SetUpMsg *Setup
}