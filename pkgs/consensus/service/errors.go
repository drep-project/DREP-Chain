package service

import "errors"

var (
	ErrMsgSize            = errors.New("err msg size")
	ErrWaitCommit         = errors.New("waitForCommit fail")
	ErrWaitResponse       = errors.New("waitForResponse fail")
	ErrSignatureNotValid  = errors.New("signature not valid")
	ErrSignBlock          = errors.New("sign block error")
	ErrBFTNotReady        = errors.New("BFT node not ready")
	ErrValidateMsg        = errors.New("validate message error")
	ErrGenerateNouncePriv = errors.New("Generate nounce fail")

	ErrTimeout       = errors.New("time out")
	ErrLowHeight     = errors.New("leader's height  lower")
	ErrHighHeight    = errors.New("leader's height  higher")
	ErrStatus        = errors.New("error status")
	ErrLeaderMistake = errors.New("setUp: mistake leader")
	ErrChallenge     = errors.New("challenge error")
)
