package service

import "errors"

var (
	ErrExistKeystore   = errors.New("exist keystore")
	ErrClosedWallet    = errors.New("wallet is not open")
	ErrLockedWallet    = errors.New("wallet is already locked")
	ErrNotAHash        = errors.New("msg is not a hash")
	ErrAlreadyUnLocked = errors.New("wallet is already unlocked")
	ErrExistKey        = errors.New("privkey is exist")
	ErrMissingKeystore        = errors.New("not found keystore")
)
