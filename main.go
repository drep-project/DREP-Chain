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
	network.Start(func(peer *network.Peer, t int, msg interface{}) {
		p := processor.GetInstance()
		if msg != nil {
			p.Process(peer, t, msg)
		}
	}, store.GetPort())
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
				if node.SendTransaction(t) != nil {
					fmt.Println("Offline")
				} else {
					fmt.Println("Send finish")
				}
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
		case "miner":
			{
				pk := store.GetPubKey()
				if pk.Equal(store.GetAdminPubKey()) {
					fmt.Print("Who: ")
					var addr string
					fmt.Scanln(&addr)
					t := node.GenerateMinerTransaction(addr)
					if node.SendTransaction(t) != nil {
						fmt.Println("Offline")
					}
				} else {
					fmt.Println("You are not allowed.")
				}
			}
		case "exit":
			{
				break
			}
		}
	}
}