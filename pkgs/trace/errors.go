package trace

import "errors"

var (
	ErrTxNotFound      = errors.New("tx not found")
	ErrBlockNotFound      = errors.New("block not found")
	ErrUnSupportDbType = errors.New("not support persistence type")
)
