package solo

import "errors"

var (
	ErrSignBlock     = errors.New("sign block error")
	ErrWalletNotOpen = errors.New("wallet is close")
	ErrCheckSigFail  = errors.New("verify sig in block fail")
	ErrGasUsed       = errors.New("gasused not match")
)
