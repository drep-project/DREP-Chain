package service

import "errors"

var (
	ErrSignBlock = errors.New("sign block error")
	ErrWalletNotOpen =  errors.New("wallet is close")
)
