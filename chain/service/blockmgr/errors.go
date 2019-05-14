package service

import (
	"errors"
)

var (

	ErrBlockNotFound             = errors.New("block not exist")
	ErrTxIndexOutOfRange         = errors.New("tx index out of range")
	ErrReachGasLimit             = errors.New("gas limit reached")
	ErrOverFlowMaxMsgSize        = errors.New("msg exceed max size")
	ErrEnoughPeer           = errors.New("peer exceed max peers")
	ErrNotContinueHeader    = errors.New("non contiguous header")
	ErrFindAncesstorTimeout = errors.New("findAncestor timeout")
	ErrGetHeaderHashTimeout = errors.New("get header hash timeout")
	ErrGetBlockTimeout      = errors.New("fetch blocks timeout")
	ErrReqStateTimeout      = errors.New("req state timeout")
	ErrDecodeMsg            = errors.New("fail to decode p2p msg")
	ErrMsgType              = errors.New("not expected msg type")
	ErrBlockExsist          = errors.New("already have block")
	ErrOrphanBlockExsist    = errors.New("already have block (orphan)")
)
