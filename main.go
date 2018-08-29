package main

import (
	"BlockChainTest/node"
	"BlockChainTest/network"
	"BlockChainTest/processor"
	"BlockChainTest/store"
	"time"
	"fmt"
	"BlockChainTest/bean"
	"math/big"
	"BlockChainTest/crypto"
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

	network.Start(func(t int, msg interface{}) {
		p := processor.GetInstance()
		if msg != nil {
			p.Process(t, msg)
		}
	})
	store.ChangeRole(role)
	processor.GetInstance().Start()
	if role == node.LEADER {
		s := store.GetItSelfOnLeader().ProcessConsensus([]byte{100, 200, 234})
		fmt.Println("Leader get sig ", s)
	} else {
		store.GetItSelfOnMember().ProcessConsensus()
	}
	time.Sleep(4000 * time.Second)
}

func main12() {
	curve := crypto.GetCurve()
	a1, _ := new(big.Int).SetString("113893772716468263639633953315095915869033802907869568852055845823400105588294", 10)
	b1, _ := new(big.Int).SetString("58178181769897769184869928777060134385665655366196217376237482262300579115473", 10)
	a2, _ := new(big.Int).SetString("48993216527568937279647182951833443538588247166304622592959962103558717114324", 10)
	b2, _ := new(big.Int).SetString("52382042545561081421374974591111006062040959972679261533597475638016134340481", 10)
	p1 := &bean.Point{X: a1.Bytes(), Y: b1.Bytes()}
	p2 := &bean.Point{X: a2.Bytes(), Y: b2.Bytes()}
	p := curve.Add(p1, p2)
	px, py := p.Int()
	fmt.Println("px: ", px)
	fmt.Println("py: ", py)
}