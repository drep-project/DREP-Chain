package pool

import (
    "BlockChainTest/util/messagepool"
    "time"
)

var (
    msgPool *messagepool.MessagePool
)

func init()  {
    msgPool = messagepool.New()
}

func ObtainMsg(num int, cp func(interface{})bool, duration time.Duration) []interface{} {
    return msgPool.Obtain(num, cp, duration)
}

func ObtainOneMsg(cp func(interface{})bool, duration time.Duration) interface{} {
    return msgPool.ObtainOne(cp, duration)
}

func PushMsg(msg interface{})  {
    msgPool.Push(msg)
}

func ContainsMsg(cp func(interface{})bool) bool {
    return msgPool.Contains(cp)
}