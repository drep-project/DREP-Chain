package types

import "github.com/drep-project/drep-chain/network/types"

const (
	MsgTypeSetUp  = 9
	MsgTypeCommitment = 10
	MsgTypeResponse = 11
	MsgTypeChallenge = 12
	MsgTypeFail = 13
)

type RouteMsgWrap struct {
	Peer *types.Peer
	SetUpMsg *Setup
}