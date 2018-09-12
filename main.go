package main

import (
	"BlockChainTest/network"
	"BlockChainTest/processor"
	"BlockChainTest/node"
	"fmt"
	"math/big"
	"BlockChainTest/bean"
	"BlockChainTest/store"
)

func main()  {
	network.Start(func(t int, msg interface{}) {
		p := processor.GetInstance()
		if msg != nil {
			p.Process(t, msg)
		}
	})
	processor.GetInstance().Start()
	node.GetNode().Start()
	for {
		var cmd string
		fmt.Scanln(&cmd)
		switch cmd {
		case "send":
			{
				var addr string
				var amount int64
				fmt.Print("To: ")
				fmt.Scanln(&addr)
				fmt.Print("Amount: ")
				fmt.Scanln(&amount)
				t := node.GenerateBalanceTransaction(bean.Address(addr), big.NewInt(amount))
				node.SendTransaction(t)
				fmt.Println("Send finish")
			}
		case "checkBalance":
			{
				var addr string
				fmt.Print("Who: ")
				fmt.Scanln(&addr)
				fmt.Println(store.GetBalance(bean.Address(addr)))
			}
		case "checkNonce":
			{
				var addr string
				fmt.Print("Who: ")
				fmt.Scanln(&addr)
				fmt.Println(store.GetNonce(bean.Address(addr)))
			}
		case "me":
			{
				addr := store.GetAddress()
				fmt.Println("Addr: ", addr)
				fmt.Println("Nonce: ", store.GetNonce(addr))
				fmt.Println("Bal: ", store.GetBalance(addr))
			}
		case "exit":
			{
				break
			}
		}
	}
}