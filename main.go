package main

import (
	"BlockChainTest/node"
	"BlockChainTest/store"
	"BlockChainTest/processor"
	"BlockChainTest/network"
)

var (
	role = node.LEADER
)
func main()  {

	//messages := make(chan *common.Message, 100)
	//net := network.GetInstance(messages)
	//if net.Start() != 0 {
	//	fmt.Errorf("error")
	//	return
	//}
	//processors := processor.GetInstance(messages)
	//processors.Start()
	//for i := 0; i < 1000; i++{
	//	time.Sleep(1 * time.Second)
	//}
	//network.ExecuteMultiSign()

	//curve := network.InitCurve()
	//prvKey, _ := network.GenerateKey(curve)
	//msg := []byte("it is rainy today，今天下雨了")
	//cipher := network.Encrypt(curve, prvKey.PubKey, msg)
	//plain := network.Decrypt(curve, prvKey, cipher)
	//if plain != nil {
	//	fmt.Println("plain: ", string(plain))
	//} else {
	//	fmt.Println("plain: ", nil)
	//}

	//test.RemoteConnect(1)
	network.Listen(func(t int, msg interface{}) {
		p := processor.GetInstance()
		if msg != nil {
			p.Process(t, msg)
		}
	})
	network.Work()
	store.ChangeRole(role)
	processor.GetInstance().Start()
	if role == node.LEADER {
		store.GetItSelfOnLeader().ProcessConsensus([]byte{100, 200, 300})
	} else {
		store.GetItSelfOnMember().ProcessConsensus()
	}
}