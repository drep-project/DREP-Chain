package service

import (
    "math/big"
)

var (
    BlockGasLimit            = big.NewInt(5000000000000000000)
)

const (
    Rewards = 1000000000000 //每出一个块，系统奖励的币数目
)

