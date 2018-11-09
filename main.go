package main

import (
	"BlockChainTest/network"
	"BlockChainTest/processor"
	"BlockChainTest/node"
	"fmt"
	"math/big"
	"BlockChainTest/bean"
	"BlockChainTest/store"
	"BlockChainTest/database"
)

func main()  {
	network.Start(func(peer *network.Peer, t int, msg interface{}) {
		p := processor.GetInstance()
		if msg != nil {
			p.Process(peer, t, msg)
		}
	}, store.GetPort())
	network.HttpStart()
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
				fmt.Println(database.GetBalance(bean.Hex2Address(addr)))
			}
		case "checkNonce":
			{
				var addr string
				fmt.Print("Who: ")
				fmt.Scanln(&addr)
				fmt.Println(database.GetNonce(bean.Hex2Address(addr)))
			}
		case "me":
			{
				addr := bean.Hex2Address(store.GetAddress().String())
				fmt.Println("Addr: ", addr)
				nonce, err := database.GetNonce(addr)
				fmt.Println("Nonce: ", nonce)
				fmt.Println("Nonce err: ", err)
				balance, err := database.GetBalance(addr)
				fmt.Println("Bal: ", balance)
				fmt.Println("Bal err: ", err)
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

//TODO (1)智能合约代码放进去(core文件夹, bean文件件里新加的account.go，account.pb.go)；接口是runtime.go里面的ApplyTransaction(*Transaction);
//TODO (2)数据库部分新加GetBlock, PutBlock, GetBalance, PutBalance等接口;
//TODO (3)哈希函数改成以太坊的SHA3算法；
//TODO (4)Block和Transaction字段填完整
//TODO (5)Http Url