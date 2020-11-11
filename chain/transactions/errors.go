package transactions

import (
	"errors"
)

var (
	ErrNonceTooHigh              = errors.New("nonce too high")
	ErrNonceTooLow               = errors.New("nonce too low")
	ErrBalance                   = errors.New("not enough balance")
	ErrInsufficientBalanceForGas = errors.New("insufficient balance to pay for gasRemained")
	ErrOutOfGas                  = errors.New("out gas of block")
	ErrTxUnSupport               = errors.New("unsupported transaction type")
)
