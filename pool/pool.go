package pool

import (
    "BlockChainTest/util/messagepool"
    "time"
)

var (
    pool *messagepool.MessagePool
)

func init()  {
    pool = messagepool.New()
}

func Obtain(num int, cp func(interface{})bool, duration time.Duration) []interface{} {
    return pool.Obtain(num, cp, duration)
}

func ObtainOne(cp func(interface{})bool, duration time.Duration) interface{} {
    return pool.ObtainOne(cp, duration)
}

func Push(msg interface{})  {
    pool.Push(msg)
}

func Contains(cp func(interface{})bool) bool {
    return pool.Contains(cp)
}