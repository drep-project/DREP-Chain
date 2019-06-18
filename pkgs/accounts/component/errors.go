package component

import "errors"

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrDecryptFail = errors.New("decryption failed")
	ErrPassword    = errors.New("password not correct")
	ErrSaveKey     = errors.New("save key failed")
	ErrDecrypt     = errors.New("could not decrypt key with given passphrase")
)
