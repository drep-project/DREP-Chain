package main

import (
	"fmt"
	"time"
	"BlockChainTest/common"
	"BlockChainTest/network"
	"BlockChainTest/processor"
)

func main()  {
	messages := make(chan *common.Message, 100)
	net := network.GetInstance(messages)
	if net.Start() != 0 {
		fmt.Errorf("error")
		return
	}
	processors := processor.GetInstance(messages)
	processors.Start()
	for i := 0; i < 1000; i++{
		time.Sleep(1 * time.Second)
	}
}