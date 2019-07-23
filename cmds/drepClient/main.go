package main

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"sync"
	"time"

	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/rpc"
	"gopkg.in/urfave/cli.v1"
)

var (
	methods = make(map[string]func(args cli.Args, client *rpc.Client, ctx context.Context))
)

func init() {
	//key是完整的方法名称
	methods["chain_getBalance"] = callGetBalance
	methods["chain_getPoolTransactions"] = callgetPoolTransactions
	methods["chain_getPoolMiniPendingNonce"] = callGetPoolMiniPendingNonce

	methods["account_createCode"] = callCreateCode
	methods["account_call"] = callContract
}

func main() {
	app := cli.NewApp()
	app.Action = action
	if err := app.Run(os.Args); err != nil {
		fmt.Println("exec err:", err)
		return
	}
}

func action(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) <= 1 {
		return fmt.Errorf("arg num too small")
	}
	//1 检查第一个字段是否是http://127.0.0.1:5555格式
	url, err := url.Parse(args[0])
	if err != nil {
		return err
	}

	fmt.Println()

	ctxReq, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := rpc.DialContext(ctxReq, url.String())
	if err != nil {
		return err
	}
	defer client.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		if fn, ok := methods[args[1]]; ok {
			fn(args[1:], client, ctxReq)
		} else {
			fmt.Println("cannot deal method:", args[0], "need register it")
		}
		wg.Done()
	}()

	wg.Wait()
	return nil
}

func callGetBalance(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := new(big.Int)
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
}

func callCreateCode(args cli.Args, client *rpc.Client, ctx context.Context) {
	r := ""
	if err := client.CallContext(ctx, &r, args[0], args[1], args[2]); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r)
}

func callContract(args cli.Args, client *rpc.Client, ctx context.Context) {
	//todo 使用通用判断代替
	if len(args) > 6 || len(args) < 4 {
		fmt.Println("param num:", len(args), "too much")
		return
	}

	sliceArgs := make([]interface{}, len(args)-1)
	for i := 0; i < len(sliceArgs); i++ {
		if i > 2 {
			value, b := new(big.Int).SetString(args[i+1], 10)
			if b == false {
				fmt.Println("param err", args[i+1])
				return
			}

			cb := new(common.Big)
			cb.SetMathBig(*value)
			//big.Int转换成json格式发出去
			buf, err := cb.MarshalText()
			if err != nil {
				fmt.Println(err)
				return
			}

			sliceArgs[i] = string(buf)

		} else {
			sliceArgs[i] = args[i+1]
		}
	}

	fmt.Println(sliceArgs)

	r := ""
	if err := client.CallContext(ctx, &r, args[0], sliceArgs...); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r)
}

func callgetPoolTransactions(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]chainTypes.Transactions, 0, 2)
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println(err)
	}

	fmt.Println("queue txs:")
	for _, tx := range resp[0] {
		fmt.Println("type:", tx.Type(), "nonce:", tx.Nonce(), "amount:", tx.Amount(), "gas:", tx.Gas(), "gasPrice:", tx.GasPrice())
	}

	fmt.Println("pending txs:")
	for _, tx := range resp[1] {
		fmt.Println("type:", tx.Type(), "nonce:", tx.Nonce(), "amount:", tx.Amount(), "gas:", tx.Gas(), "gasPrice:", tx.GasPrice())
	}
}

func callGetPoolMiniPendingNonce(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp uint64
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
}
