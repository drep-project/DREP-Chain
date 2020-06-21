package main

import (
	"errors"
	"fmt"
	"log"
	"testing"
)

func TestFuncParser(t *testing.T) {
	example := `
 name:获取地址
 usage:  用于获取区块信息
 params: 1: addr string  2: addr string 3: addr string  4: addr string
return: 所有账户hash地址的数组
 example: curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getNonce","params":["0xc4ac59f52b3052e5c14566ed397453ea913c6fbc"], "id": 3}' -H "Content-Type:application/json"
 response:
{
	"jsonrpc": "2.0",
	"id": 1,
	"result": ["0x0e4b43c04e4b9a57cf80fd8aad70960fca004af7", "0x76d737c4aac27b7c8a56d8566b28a307d821cd99", "0x8228385f1580551ef5c38fd5e84986f3ef57c196", "0xc4ac59f52b3052e5c14566ed397453ea913c6fbc"]
}
`
	tokens := funcParser(example, "")
	if tokens.Name != "获取地址" {
		log.Fatal(errors.New("parser func err"))
	}
	fmt.Println(tokens)
}

func TestStructParser(t *testing.T) {
	example := `
name: 链
 usage:  用于获取区块信息
`
	tokens := structParser(example)
	if tokens.Name != "链" {
		log.Fatal(errors.New("parser struct name err"))
	}
	if tokens.Tokens["usage:"].Str != "用于获取区块信息" {
		log.Fatal(errors.New("parser struct usage err"))
	}
}
