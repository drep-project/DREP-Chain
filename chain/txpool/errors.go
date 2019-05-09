package txpool

import "errors"

var (
	ErrQueueFull  = errors.New("queue full")
	ErrTxExist    = errors.New("transaction exists")
	ErrTxPoolFull = errors.New("transaction pool full")
)
