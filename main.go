package main

import (
	"BlockChainTest/node"
	"BlockChainTest/network"
	"BlockChainTest/processor"
	"time"
	"BlockChainTest/bean"
	"BlockChainTest/store"
)

var (
	role = bean.LEADER
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

	network.Start(func(t int, msg interface{}) {
		p := processor.GetInstance()
		if msg != nil {
			p.Process(t, msg)
		}
	})
	processor.GetInstance().Start()
	node.NewNode(store.GetPrvKey()).Start()
	//if role == node.LEADER {
	//	s, b := store.GetItSelfOnLeader().ProcessConsensus([]byte{100, 200, 234})
	//	log.Println("Leader get sig ", s)
	//	log.Println("Leader get bytes ", b)
	//} else {
	//	store.GetItSelfOnMember().ProcessConsensus()
	//}
	time.Sleep(4000 * time.Second)
}