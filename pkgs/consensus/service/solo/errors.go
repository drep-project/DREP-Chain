package solo

import "errors"

var (
	ErrCheckSigFail = errors.New("verify sig in block fail")
	ErrGasUsed = errors.New("gasused not match")
)
