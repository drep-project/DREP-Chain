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
			"0xe0b656c5763e79a201ac5b12e61a09f7c83f8183",
			"0xe0b656c5763e79a201ac5b12e61a09f7c83f8183",
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
