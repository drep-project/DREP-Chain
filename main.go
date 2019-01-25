package main

import (
	"BlockChainTest/network"
	"BlockChainTest/processor"
	"BlockChainTest/node"
	"fmt"
	"math/big"
	"BlockChainTest/store"
	"BlockChainTest/database"
	"BlockChainTest/http"
	"time"
	"BlockChainTest/accounts"
	"encoding/hex"
	"BlockChainTest/bean"
)

func main()  {
	network.Start(func(peer *bean.Peer, t int, msg interface{}) {
		p := processor.GetInstance()
		if msg != nil {
			p.Process(peer, t, msg)
		}
	}, store.GetPort())
	http.Start()
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
				var destChain int64
				fmt.Print("To: ")
				fmt.Scanln(&addr)
				fmt.Print("DestChain: ")
				fmt.Scanln(&destChain)
				fmt.Print("Amount: ")
				fmt.Scanln(&amount)
				t := node.GenerateBalanceTransaction(addr, destChain, big.NewInt(amount))
				if node.SendTransaction(t) != nil {
					fmt.Println("Offline")
				} else {
					fmt.Println("Send finish")
				}
			}
		case "checkBalance":
			{
				var addr string
				chainId := store.GetChainId()
				fmt.Print("Who: ")
				fmt.Scanln(&addr)
				fmt.Println(database.GetBalanceOutsideTransaction(accounts.Hex2Address(addr), chainId))
			}
		case "checkNonce":
			{
				var addr string
				chainId := store.GetChainId()
				fmt.Print("Who: ")
				fmt.Scanln(&addr)
				fmt.Println(database.GetNonceOutsideTransaction(accounts.Hex2Address(addr), chainId))
			}
		case "me":
			{
				addr := store.GetAddress()
				chainId := store.GetChainId()
				fmt.Println("Addr: ", addr.Hex())
				nonce := database.GetNonceOutsideTransaction(addr, chainId)
				fmt.Println("Nonce: ", nonce)
				balance := database.GetBalanceOutsideTransaction(addr, chainId)
				fmt.Println("Bal: ", balance)
			}
		case "miner":
			{
				pk := store.GetPubKey()
				if pk.Equal(store.GetAdminPubKey()) {
					fmt.Print("Who: ")
					var addr string
					chainId := store.GetChainId()
					fmt.Scanln(&addr)
					t := node.GenerateMinerTransaction(addr, chainId)
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
		case "create":
			{
				var code string
				fmt.Print("code: ")
				fmt.Scanln(&code)
				byt, _ := hex.DecodeString(code)
				t := node.GenerateCreateContractTransaction(byt)
				node.SendTransaction(t)
			}
		case "call":
			{
				var addr string
				var chainId int64
				var input string
				var readOnly bool
				fmt.Print("addr: ")
				fmt.Scanln(&addr)
				fmt.Print("chainId: ")
				fmt.Scanln(&chainId)
				fmt.Print("input: ")
				fmt.Scanln(&input)
				fmt.Print("readOnly: ")
				fmt.Scanln(&readOnly)
				inp, _ := hex.DecodeString(input)
				t := node.GenerateCallContractTransaction(accounts.Hex2Address(addr), chainId, inp, readOnly)
				node.SendTransaction(t)
			}
		case "check":
			{
				var addr string
				var chainId int64
				fmt.Print("addr: ")
				fmt.Scanln(&addr)
				fmt.Print("chainId: ")
				fmt.Scanln(&chainId)
				storage := database.GetStorageOutsideTransaction(accounts.Hex2Address(addr), chainId)
				fmt.Println("contract storage: ", storage)
			}
		case "do":
			{
				itr := database.GetItr()
				fmt.Println()
				fmt.Println()
				for itr.Next() {
					key := itr.Key()
					value := itr.Value()
					if len(string(value)) > 100 && len(string(value)) < 500 {
						fmt.Println(hex.EncodeToString(key), ": ", string(value))
					}
					//if len(string(value)) > 500 {
					//	//fmt.Println(hex.EncodeToString(key), ": ", string(value))
					//	fmt.Println("max height: ", database.GetMaxHeight())
					//}
				}
				fmt.Println()
				fmt.Println()
			}
		}
	}
}

func main1() {
	http.Start()
	time.Sleep(3600 * time.Second)
}

//TODO (1)智能合约代码放进去(core文件夹, bean文件件里新加的account.go，accounts.pb.go)；接口是runtime.go里面的ApplyTransaction(*transaction);
//TODO (2)数据库部分新加GetBlock, PutBlock, GetBalance, PutBalance等接口;
//TODO (3)哈希函数改成以太坊的SHA3算法；
//TODO (4)Block和Transaction字段填完整
//TODO (5)Http Url