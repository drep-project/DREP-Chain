package blockmgr

import (
	"errors"
)

var (
	// ErrBlockNotFound print error message.
	ErrBlockNotFound = errors.New("block not exist")
	// ErrTxIndexOutOfRange print error message.
	ErrTxIndexOutOfRange = errors.New("tx index out of range")
	// ErrReachGasLimit print error message.
	ErrReachGasLimit = errors.New("gas limit reached")
	// ErrOverFlowMaxMsgSize print error message.
	ErrOverFlowMaxMsgSize = errors.New("msg exceed max size")
	// ErrEnoughPeer print error message.
	ErrEnoughPeer = errors.New("peer exceed max peers")
	// ErrNotContinueHeader print error message.
	ErrNotContinueHeader = errors.New("non contiguous header")
	// ErrFindAncesstorTimeout print error message.
	ErrFindAncesstorTimeout = errors.New("findAncestor timeout")
	// ErrGetHeaderHashTimeout print error message.
	ErrGetHeaderHashTimeout = errors.New("get header hash timeout")
	// ErrGetBlockTimeout print error message.
	ErrGetBlockTimeout = errors.New("fetch blocks timeout")
	// ErrReqStateTimeout print error message.
	ErrReqStateTimeout = errors.New("req state timeout")
	// ErrDecodeMsg print error message.
	ErrDecodeMsg = errors.New("fail to decode p2p msg")
	// ErrMsgType print error message.
	ErrMsgType = errors.New("not expected msg type")
	// ErrNegativeAmount print error message.
	ErrNegativeAmount = errors.New("negative amount in tx")
	// ErrExceedGasLimit print error message.
	ErrExceedGasLimit = errors.New("gas limit in tx has exceed block limit")
	// ErrBalance print error message.
	ErrBalance = errors.New("not enough balance")
	// ErrNotSupportRenameAlias print error message.
	ErrNotSupportRenameAlias = errors.New("not suppport rename alias")
	// ErrNoCommonAncesstor print error message.
	ErrNoCommonAncesstor = errors.New("no common ancesstor")
)
