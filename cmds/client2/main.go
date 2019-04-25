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
			"0x09926e07b4dd2a38c497da613c008ede1e2b1506",
			"0x09926e07b4dd2a38c497da613c008ede1e2b1506",
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
