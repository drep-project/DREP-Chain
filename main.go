package main
import (
	"BlockChainTest/network"
	"BlockChainTest/common"
	"fmt"
	"BlockChainTest/processor"
	"time"
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
	//network.ExecuteMultiSign()
}