package main

import (
	"BlockChainTest/network"
	"BlockChainTest/processor"
	"BlockChainTest/node"
	"BlockChainTest/store"
	"time"
)

func main()  {
	network.Start(func(t int, msg interface{}) {
		p := processor.GetInstance()
		if msg != nil {
			p.Process(t, msg)
		}
	})
	processor.GetInstance().Start()
	node.GetNode(store.GetPrvKey()).Start()
	time.Sleep(4000 * time.Second)

}