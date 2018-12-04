package main 

import (
	"fmt"
	"strconv"
	"encoding/json"
	"os"
	"strings"
	"BlockChainTest/rpc"
	"github.com/ethereum/go-ethereum/log"
)

func main(){
	var (
		client *rpc.Client
		err    error
	)
	// Attach to an Ethereum node over IPC or RPC   "http://127.0.0.1:15645"//
	endpoint :=  "http://127.0.0.1:15645"////"rpc:http://"+rpc.DefaultHTTPEndpoint()
	if client, err = dialRPC(endpoint); err != nil {
		log.Error("unanble to connect rpc server")
		return
	}
	defer client.Close()

	if len(os.Args) <2 {
		log.Error("argument format error. eg: getblock 1")
	}
	result := &json.RawMessage{}
	methodName := os.Args[1]
	args := os.Args[2:]
    var iargs []interface{}
	for _, arg := range args {
		num, err := strconv.ParseInt(arg, 10, 64)
		if err !=nil {
			iargs =append(iargs,arg)
		}else{
			iargs =append(iargs,num)
		}
	}

	err = client.Call(result, methodName,iargs...)
	if err != nil {
		fmt.Println(err.Error())
		log.Error(err.Error())
	}else {

		bytes, _ := json.MarshalIndent(result, "", "\t")
		fmt.Println(string(bytes))
	}
}

// dialRPC returns a RPC client which connects to the given endpoint.
// The check for empty endpoint implements the defaulting logic
// for "geth attach" and "geth monitor" with no argument.
func dialRPC(endpoint string) (*rpc.Client, error) {
	if strings.HasPrefix(endpoint, "rpc:") || strings.HasPrefix(endpoint, "ipc:") {
		// Backwards compatibility with geth < 1.5 which required
		// these prefixes.
		endpoint = endpoint[4:]
	}
	return rpc.Dial(endpoint)
}
