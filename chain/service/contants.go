package service

import (
	"math/big"
)

var (
	BlockGasLimit = big.NewInt(5000000000000000000)
)

const (
	maxHeaderHashCountReq = 255                 //最多一次请求的头部hash个数
	maxBlockCountReq      = 16                  //最多一次请求的头部hash个数
	maxPeerCountReq       = 3                   //每次请求最大发给peer的数目
	maxSyncSleepTime      = 200                 //同步的过程中，每个周期休息200毫秒
	maxNetworkTimeout     = 5                   //最大网络超时时间
	Rewards               = 1000000000000000000 //每出一个块，系统奖励的币数目
)
