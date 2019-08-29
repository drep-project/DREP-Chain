package bft

import "errors"

var (
	ErrBpConfig           = errors.New("the pubkey config not in bp nodes")
	ErrBFTNotReady        = errors.New("BFT node not ready")
	ErrBpNotInList        = errors.New("bp node not in local list")
	ErrMultiSig           = errors.New("ErrMultiSig")
	ErrWaitCommit         = errors.New("waitForCommit fail")
	ErrWaitResponse       = errors.New("waitForResponse fail")
	ErrChallenge          = errors.New("challenge error")
	ErrSignatureNotValid  = errors.New("signature not valid")
	ErrTimeout            = errors.New("time out")
	ErrLowHeight          = errors.New("leader's height  lower")
	ErrHighHeight         = errors.New("leader's height  higher")
	ErrStatus             = errors.New("error status")
	ErrLeaderMistake      = errors.New("setUp: mistake leader")
	ErrValidateMsg        = errors.New("validate message error")
	ErrGenerateNouncePriv = errors.New("Generate nounce fail")
	ErrMsgSize            = errors.New("err msg size")
	ErrGasUsed            = errors.New("GasUsed not match gasUsed in blockheader")
)
