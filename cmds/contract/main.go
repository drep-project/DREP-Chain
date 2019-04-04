package main

import (
	"fmt"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/rpc"
	"math/big"
)

const (
	rawUrl = "http://localhost:15645"
)
var (
	pathFlag = common.DirectoryFlag{
		Name:  "path",
		Usage: "keystore save to",
	}
)

func main() {
	balanceOf("dreptest1", "gg7")
}

func balanceOf(account string ,contract string) error {
	client, err := rpc.Dial(rawUrl)
	if err != nil {
		return err
	}
	var result interface{}
	sink := common.ZeroCopySink{}
	//sink.WriteString("transfer")
	//sink.WriteString("dreptest1")
	//sink.WriteString("account1")
	//sink.WriteU256(big.NewInt(111))

	sink.WriteString("balanceOf3")
	val := new (big.Int)
	val.SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	val.Add(val, big.NewInt(2))
	sink.WriteU256(val)
	err = client.Call(&result, "account_call", account, contract, sink.Bytes())
	fmt.Println(err)
	returnVal := result.(map[string]interface{})["Return"].(string)
	TxHash := result.(map[string]interface{})["TxHash"].(string)
	fmt.Println(TxHash)

	returnBytes, _:=common.Decode(returnVal)
	source := common.NewZeroCopySource(returnBytes)
	balance,_ := source.Nextu256()
	fmt.Println(balance)
	return nil
}

func transfer(from, to string, amount *big.Int, contract string) error{
	client, err := rpc.Dial(rawUrl)
	if err != nil {
		return err
	}

	sink := common.ZeroCopySink{}
	sink.WriteString("transfer")
	sink.WriteString(from)
	sink.WriteString(to)
	sink.WriteU256(amount)
	var result interface{}
	err = client.Call(&result, "account_call", contract,from, to, sink.Bytes())
	fmt.Println(result)
	return nil
}

func totalSupply(contract string) error{
	client, err := rpc.Dial(rawUrl)
	if err != nil {
		return err
	}

	sink := common.ZeroCopySink{}
	sink.WriteString("symbol")

	var result interface{}
	err = client.Call(&result, "account_call", contract, sink.Bytes())

	returnVal := result.(map[string]interface{})["Return"].(string)
	returnBytes, _:=common.Decode(returnVal)
	source := common.NewZeroCopySource(returnBytes)
	fmt.Println(source.Nextu256())
	return nil
}

func symbol(contract string) error{
	client, err := rpc.Dial(rawUrl)
	if err != nil {
		return err
	}

	sink := common.ZeroCopySink{}
	sink.WriteString("symbol")

	var result interface{}
	err = client.Call(&result, "account_call", contract, sink.Bytes())

	returnVal := result.(map[string]interface{})["Return"].(string)
	returnBytes, _:=common.Decode(returnVal)
	source := common.NewZeroCopySource(returnBytes)
	fmt.Println(source.NextString())
	return nil
}

func name(contract string) error{
	client, err := rpc.Dial(rawUrl)
	if err != nil {
		return err
	}

	sink := common.ZeroCopySink{}
	sink.WriteString("name")

	var result interface{}
	err = client.Call(&result, "account_call", contract, sink.Bytes())

	returnVal := result.(map[string]interface{})["Return"].(string)
	returnBytes, _:=common.Decode(returnVal)
	source := common.NewZeroCopySource(returnBytes)
	fmt.Println(source.NextString())
	return nil
}

func initContract(contract string) error{
	client, err := rpc.Dial(rawUrl)
	if err != nil {
		return err
	}

	sink := common.ZeroCopySink{}
	sink.WriteString("init")

	var result interface{}
	err = client.Call(&result, "account_call", contract, sink.Bytes())
	fmt.Println(err)
	fmt.Println(result)
	return nil
}

func Create(caller string ,contract string, code []byte) error {
	client, err := rpc.Dial(rawUrl)
	if err != nil {
		return err
	}

	var result interface{}
	err = client.Call(result, "account_createContract", caller, contract, code)
	fmt.Println(err)
	fmt.Println(result)
	return nil
}