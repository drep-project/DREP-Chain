package main

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/rpc"
	"io/ioutil"
	"log"
	"math/big"
)

var (
	client *rpc.Client
)
func main(){
	content, err := ioutil.ReadFile("D:\\goproject\\src\\github.com\\drep-project\\drep-chain\\cmds\\client\\config.json")
	if err != nil {
		log.Fatal(err)
	}
	config := &Config{}
	err = json.Unmarshal(content, config)
	if err != nil {
		log.Fatal(err)
	}
//	url := "http://127.0.0.1:15645"//args[0]

	client, err = rpc.DialHTTP(config.Url)
	if err != nil {
		panic(err)
	}
	err = run(config.Actions)
	if err != nil {
		log.Fatal(err)
	}
}

func run(actions []Action)  error {
	for _, action := range  actions {
		var err error
		if action.Method == "transfer" {
			err = ResolveTransfer(action)
		}else if action.Method == "alias" {
			err = ResolveSetAlias(action)
		}
		if err != nil {
			return  err
		}
	}
	return nil
}

func ResolveTransfer(action Action) error {
	result := ""
	args := []interface{}{}
	transferArgs := &TransArgs{}
	json.Unmarshal(action.Args, transferArgs)


	args = append(args, transferArgs.From)
	args = append(args, transferArgs.To)
	args = append(args, NumberToHex(transferArgs.Amount))
	args = append(args, NumberToHex(transferArgs.GasPrice))
	args = append(args, NumberToHex(transferArgs.GasLimit))
	args = append(args, common.Encode([]byte(transferArgs.Data)))
	err := client.Call(&result,"account_transfer",args...)
	/*
	err := client.Call(&result,"account_transfer",
		"0xe91f67944ec2f7223bf6d0272557a5b13ecc1f28",
		"0xe91f67944ec2f7223bf6d0272557a5b13ecc1f28",
		"0x3e8",
		"0x7530",
		"0x7530",
		"0x3b9aca00",
	)
	*/
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func ResolveSetAlias(action Action) error {
	result := ""
	args := []interface{}{}
	setAliasArgs := &SetAliasArgs{}
	json.Unmarshal(action.Args, setAliasArgs)

	args = append(args, setAliasArgs.From)
	args = append(args, setAliasArgs.Alias)
	args = append(args, NumberToHex(setAliasArgs.GasPrice))
	args = append(args, NumberToHex(setAliasArgs.GasLimit))
	err := client.Call(&result,"account_setAlias",args...)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func NumberToHex(number string)string{
	val := new (big.Int)
	val.SetString(number, 10)
	return common.EncodeUint64(val.Uint64())
}
type Config struct {
	Url string
	Actions []Action
}

type Action struct {
	Method string
	Args json.RawMessage
}

type TransArgs struct {
	From   string
	To     string
	Amount string
	GasPrice string
	GasLimit string
	Data     string
}

type SetAliasArgs struct {
	From   string
	Alias     string
	GasPrice string
	GasLimit string
}