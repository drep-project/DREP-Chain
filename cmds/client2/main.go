package main

import (
	"fmt"
	"github.com/drep-project/drep-chain/rpc"
)

func main(){
//	args := os.Args
	url := "http://127.0.0.1:15645"//args[0]
	stopChanel := make(chan struct{})

	client, err := rpc.DialHTTP(url)
	if err != nil {
		panic(err)
	}

	for {
		result := ""
		err = client.Call(&result,"account_transfer",
			"0xe91f67944ec2f7223bf6d0272557a5b13ecc1f28",
			"0xe91f67944ec2f7223bf6d0272557a5b13ecc1f28",
			"0x3e8",
			"0x7530",
			"0x7530",
			"0x3b9aca00",
		)
		fmt.Println(err)
		fmt.Println(result)
	}

	<-stopChanel
}
