package component

import "errors"

var (
	ErrKeyNotExistOrUnlock = errors.New("key not exit or not unlock account")
	ErrKeyNotFound         = errors.New("key not found")
	ErrDecryptFail         = errors.New("decryption failed")
	ErrPassword            = errors.New("password not correct")
	ErrSaveKey             = errors.New("save key failed")
	ErrDecrypt             = errors.New("could not decrypt key with given passphrase")
	ErrLocked			   = errors.New("account locked")
)
