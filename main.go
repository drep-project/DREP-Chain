package main

import (
	"time"
	"BlockChainTest/test"
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

	test.RemoteConnect(1)
	time.Sleep(3600 * time.Second)
}