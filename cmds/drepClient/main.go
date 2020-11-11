package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/common/hexutil"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/network/p2p/enode"
	p2pTypes "github.com/drep-project/DREP-Chain/network/types"
	"github.com/drep-project/DREP-Chain/params"
	accountComponent "github.com/drep-project/DREP-Chain/pkgs/accounts/component"
	"github.com/drep-project/DREP-Chain/pkgs/accounts/service"
	accountTypes "github.com/drep-project/DREP-Chain/pkgs/accounts/types"
	chainIndexerTypes "github.com/drep-project/DREP-Chain/pkgs/chain_indexer"
	cservice "github.com/drep-project/DREP-Chain/pkgs/consensus/service"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service/bft"
	filterTypes "github.com/drep-project/DREP-Chain/pkgs/filter"
	"github.com/drep-project/DREP-Chain/pkgs/log"
	"github.com/drep-project/DREP-Chain/pkgs/trace"
	"math/big"
	"net/url"
	"os"
	path2 "path"
	"strconv"
	"sync"
	"time"

	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/rpc"
	"gopkg.in/urfave/cli.v1"
)

var (
	methods = make(map[string]func(args cli.Args, client *rpc.Client, ctx context.Context))
)

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
	if len(args) < 1 {
		return fmt.Errorf("please input request url")
	}
	if len(args) == 1 {
		return fmt.Errorf("please input request method")
	}
	//1 Check whether the first field like :http://127.0.0.1:5555
	url, err := url.Parse("http://" + args[0])
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
			fmt.Println("cannot deal method:", args[1], "need register it")
		}
		wg.Done()
	}()

	wg.Wait()
	return nil
}

func init() {
	methods["genaccount"] = genaccount
	methods["blockmgr_sendRawTransaction"] = sendRawTransaction
	methods["blockmgr_gasPrice"] = gasPrice
	methods["blockmgr_getPoolTransactions"] = getPoolTransactions
	methods["blockmgr_getTransactionCount"] = getTransactionCount
	methods["blockmgr_getPoolMiniPendingNonce"] = getPoolMiniPendingNonce
	methods["blockmgr_getTxInPool"] = getTxInPool
	methods["chain_getBlock"] = getBlock
	methods["chain_getMaxHeight"] = getMaxHeight
	methods["chain_getBlockGasInfo"] = getBlockGasInfo
	methods["chain_getBalance"] = getBalance
	methods["chain_getNonce"] = getNonce
	methods["chain_getReputation"] = getReputation
	methods["chain_getTransactionByBlockHeightAndIndex"] = getTransactionByBlockHeightAndIndex
	methods["chain_getAliasByAddress"] = getAliasByAddress
	methods["chain_getAddressByAlias"] = getAddressByAlias
	methods["chain_getReceipt"] = getReceipt
	methods["chain_getLogs"] = getLogs
	methods["chain_getCancelCreditDetailByTXHash"] = getCancelCreditDetailByTXHash
	methods["chain_getByteCode"] = getByteCode
	methods["chain_getCreditDetails"] = getCreditDetails
	methods["chain_getCancelCreditDetails"] = getCancelCreditDetails
	methods["chain_getCandidateAddrs"] = getCandidateAddrs
	methods["chain_getChangeCycle"] = getChangeCycle
	methods["p2p_getPeers"] = getPeers
	methods["p2p_addPeer"] = addPeer
	methods["p2p_removePeer"] = removePeer
	methods["log_setLevel"] = setLevel
	methods["log_setVmodule"] = setVmodule
	methods["trace_getRawTransaction"] = getRawTransaction
	methods["trace_getTransaction"] = getTransaction
	methods["trace_decodeTrasnaction"] = decodeTrasnaction
	methods["trace_getSendTransactionByAddr"] = getSendTransactionByAddr
	methods["trace_getReceiveTransactionByAddr"] = getReceiveTransactionByAddr
	methods["trace_rebuild"] = rebuild
	methods["account_listAddress"] = listAddress
	methods["account_createAccount"] = createAccount
	methods["account_createWallet"] = createWallet
	methods["account_lockAccount"] = lockAccount
	methods["account_unlockAccount"] = unlockAccount
	methods["account_openWallet"] = openWallet
	methods["account_closeWallet"] = closeWallet
	methods["account_transfer"] = transfer
	methods["account_transferWithNonce"] = transferWithNonce
	methods["account_setAlias"] = setAlias

	methods["account_voteCredit"] = voteCredit
	methods["account_cancelVoteCredit"] = cancelVoteCredit
	methods["account_candidateCredit"] = candidateCredit
	methods["account_cancelCandidateCredit"] = cancelCandidateCredit

	methods["account_readContract"] = readContract
	methods["account_estimateGas"] = estimateGas
	methods["account_executeContract"] = executeContract
	methods["account_createCode"] = createCode
	methods["account_dumpPrivkey"] = dumpPrivkey
	methods["account_dumpPubkey"] = dumpPubkey
	methods["account_sign"] = sign
	methods["account_generateAddresses"] = generateAddresses
	methods["account_importKeyStore"] = importKeyStore
	methods["account_importPrivkey"] = importPrivkey
	methods["account_getKeyStores"] = getKeyStores
	methods["consensus_changeWaitTime"] = changeWaitTime
	methods["consensus_getMiners"] = getMiners
}

func argsJudge(args cli.Args, count int) error {
	if len(args) != count {
		errStr := "parameters number: " + strconv.Itoa(len(args)) + ", unequal to required number: " + strconv.Itoa(count)
		return fmt.Errorf(errStr)
	}
	return nil
}

func candidateCredit(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 7); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5], args[6]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func sendRawTransaction(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func gasPrice(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := new(big.Int)
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func getTransactionCount(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp uint64
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getPoolMiniPendingNonce(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp uint64
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getTxInPool(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Transaction
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getBlock(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Block
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	blockNum, _ := strconv.Atoi(args[1])
	if err := client.CallContext(ctx, &resp, args[0], blockNum); err != nil {
		fmt.Println("return err :", err)
		return
	}
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getMaxHeight(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp uint64
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getBlockGasInfo(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func getNonce(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp uint64
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getReputation(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := new(big.Int)
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getTransactionByBlockHeightAndIndex(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Transaction
	if err := argsJudge(args, 3); err != nil {
		fmt.Println(err.Error())
		return
	}
	blockNum, _ := strconv.Atoi(args[1])
	index, _ := strconv.Atoi(args[2])
	if err := client.CallContext(ctx, &resp, args[0], blockNum, index); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))

}
func getAliasByAddress(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getAddressByAlias(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getReceipt(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Receipt
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}

func getLogs(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]*types.Log, 0)
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getCancelCreditDetailByTXHash(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]*types.CancelCreditDetail, 0)
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getByteCode(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp hexutil.Bytes
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp.String())
}
func getCreditDetails(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func getCancelCreditDetails(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getCandidateAddrs(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getChangeCycle(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp int
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
	}
	fmt.Println(resp)
}
func getPeers(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]string, 0)
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func addPeer(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp interface{}
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
	}
	fmt.Println("success")
}
func removePeer(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp interface{}
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func setLevel(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp interface{}
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func setVmodule(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func getRawTransaction(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getTransaction(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Transaction
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func decodeTrasnaction(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp trace.RpcTransaction
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getSendTransactionByAddr(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]*trace.RpcTransaction, 0)
	if err := argsJudge(args, 4); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getReceiveTransactionByAddr(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]*trace.RpcTransaction, 0)
	if err := argsJudge(args, 4); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func rebuild(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := argsJudge(args, 3); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func listAddress(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]string, 0)
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func createAccount(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func createWallet(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func lockAccount(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("lock success")
}
func unlockAccount(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := argsJudge(args, 3); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("unlock success")
}
func openWallet(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func closeWallet(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func transfer(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 7); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5], args[6]); err != nil {
		fmt.Println("return err :", err)
	}
	fmt.Println(resp)
}
func transferWithNonce(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 8); err != nil {
		fmt.Println(err.Error())
		return
	}
	oldNonce, _ := strconv.Atoi(args[7])
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5], args[6], oldNonce); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func setAlias(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string //error
	if err := argsJudge(args, 5); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func voteCredit(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string //(string, error)
	if err := argsJudge(args, 6); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func cancelVoteCredit(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string //error
	if err := argsJudge(args, 6); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func cancelCandidateCredit(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 6); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func readContract(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp common.Bytes //common.Bytes, error
	if err := argsJudge(args, 4); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func estimateGas(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp uint64 //(uint64, error) {
	if err := argsJudge(args, 5); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], common.Bytes(args[3]), args[4]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func dumpPrivkey(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp secp256k1.PrivateKey //*secp256k1.PrivateKey, error
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	fmt.Println("0x"+common.Bytes2Hex(resp.Serialize()))
}
func dumpPubkey(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp secp256k1.PublicKey //*secp256k1.PublicKey, error
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	fmt.Println("0x" + common.Bytes2Hex(resp.Serialize()))
}
func sign(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp common.Bytes //(common.Bytes, error)
	if err := argsJudge(args, 3); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func generateAddresses(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp service.RpcAddresses //(*RpcAddresses, error)
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func importKeyStore(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]string, 0) //([]string, error)
	if err := argsJudge(args, 3); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func importPrivkey(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string // (string, error)
	if err := argsJudge(args, 3); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getKeyStores(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func changeWaitTime(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	waitTime,_:= strconv.Atoi(args[1])
	if err := client.CallContext(ctx, &resp, args[0], waitTime); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getMiners(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]*secp256k1.PublicKey, 0) //[]*secp256k1.PublicKey
	if err := client.CallContext(ctx, &resp, args[0]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes,_:= json.Marshal(resp)
	fmt.Println(string(bytes))
}

func getBalance(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
	}

	fmt.Println(resp)
}

func createCode(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 5); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func executeContract(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := argsJudge(args, 6); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func getPoolTransactions(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]types.Transactions, 0, 2)
	if err := argsJudge(args, 2); err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
	}

	for _, txs := range resp {
		for _, tx := range txs {

			from, _ := tx.From()
			fmt.Println("from:", from.String(), "to:", tx.To().String(), "nonce:", tx.Nonce(), "amount:", tx.Amount())
			fmt.Println("txHash:", tx.TxHash())
			fmt.Println("no sign data:", hexutil.Encode(tx.AsSignMessage()))
			fmt.Println()
		}
	}
}

type Genesis struct {
	Preminer []*chain.Preminer
	Miners   []types.CandidateData
}

func getCurPath() string {
	dir, _ := os.Getwd()
	return dir
}

func genaccount(args cli.Args, client *rpc.Client, ctx context.Context)  {

	if err := argsJudge(args, 4); err != nil {
		fmt.Println("genaccount need another name and password parameters")
		fmt.Println(err.Error())
		return
	}

	path := getCurPath()
	// args[1] define name of directory
	userDir := path2.Join(path, args[1])
	// args[2] define password of keystore
	password := args[2]

	var genesis Genesis

	staticConfigs := []string{"enode://f57881c48aaccf97485c2b65b421bfeda22cc3b427c44be7607b122fc1688abb@172.104.123.143:10086", "enode://9d25d161ae4b676e2df55accca93c3137df3166326d04420ffbdf66e887bd494@172.104.116.219:10086", "enode://bc7ca1b57175f2d5c85da73d367408529468a034b97d083aaecf88196090e245@172.105.103.59:10086", "enode://0ebd0422ca32d70292be128342f9e5ca32ab3cef28dc32cc332169e578e7b4f5@109.74.203.199:10086"}
	genesisConfig := "{\"Preminer\": [{\"Addr\": \"0xfb6711cbafbd5e75c612d69db9025f7eb5096d46\", \"Value\": 10000000000000000000000000000}],\"Miners\": [{\"Pubkey\": \"0x0305bfc35e079ca7ae7ad6fcf246f8ed0247d0fdbda0b7daa10f2b2f3a88e9fd8a\",\"Node\": \"enode://f57881c48aaccf97485c2b65b421bfeda22cc3b427c44be7607b122fc1688abb@172.104.123.143:10086\"},{\"Pubkey\": \"0x0354d5d560039693ec4dedb416354bdaa7aa70823fd5fabf56ceeb76386efc5670\",\"Node\": \"enode://9d25d161ae4b676e2df55accca93c3137df3166326d04420ffbdf66e887bd494@172.104.116.219:10086\"},{\"Pubkey\": \"0x031283a05887e8b291fd909a6302cdd84b0b50bbc0ae5da8ce186435a1c6b8d10a\",\"Node\": \"enode://bc7ca1b57175f2d5c85da73d367408529468a034b97d083aaecf88196090e245@172.105.103.59:10086\"},{\"Pubkey\": \"0x0349028fc4ce0ed5b6d496fcbd12755779954c1a17e74dfb283190e14ee5ea3170\",\"Node\": \"enode://0ebd0422ca32d70292be128342f9e5ca32ab3cef28dc32cc332169e578e7b4f5@109.74.203.199:10086\"},{\"Pubkey\": \"0x03902b01673cc0ffad466f68a7a1494c9da1a80fb63dae8c9d18dfc8f9aeff1eba\",\"Node\": \"enode://c01ff36a9914f781a058c77e98f09eedd5ad7e2e575c7e6c2d6cc86d859693f2@139.162.161.8:10086\"},{\"Pubkey\": \"0x028b2388f1cd0b8056a3ef41a7559952a0abe40ce8d93f586c7aa23b02f841ed23\",\"Node\": \"enode://bb7925d62126c1058f7ad951e862aa4fc0aaaa4dc7cae2fcac3b86aee316bc02@172.104.178.138:10086\"},{\"Pubkey\": \"0x0284454961eb9c389bf292e1d9a1eb957e4b22249771ce942a9eda1768fe607b32\",\"Node\": \"enode://9863364f265843bbb2d7810c9495e83b04cc541bedc8ae44ca63a2c5ec9d1b75@172.105.182.178:10086\"},{\"Pubkey\": \"0x02e42e247ae1f9a737aaa974f83dc8ca2e96c9e9a635e7e4d124df633a886d5e0e\",\"Node\": \"enode://7523f570a35d58841440a792b51ae0608a1c7f387d7f0899c59720a5f851604a@172.105.42.46:10086\"},{\"Pubkey\": \"0x036b07d786bd104b83069049be0f918c9b1505f5a90ab645837794a20c63f2b6f1\",\"Node\": \"enode://179838dc7bdb7285f82ec8e15515e09208a7d7452898dd4b978df88b694402d3@96.126.122.90:10086\"},{\"Pubkey\": \"0x0299e438ccbca1a492dfa04ab3189b687f6311fc99cd9f6112b23b4433e4c758ec\",\"Node\": \"enode://ab789802b04057655247b51fb58f48643a7a4ae883a786ee4871b67c6e57250f@45.79.252.45:10086\"},{\"Pubkey\": \"0x025d404222e2f67a77963ae7cd846bc95174e0150d47c61665b5a280fbef989c64\",\"Node\": \"enode://2a87c3c98b416f22d11e8952b36c220942461ccd69ecdbc54aef1b0d90238da4@45.56.95.177:10086\"},{\"Pubkey\": \"0x02ea32b18f6a37b23ebccca1ed6ae4f513da764f85f3a37234c18cad627e820dc9\",\"Node\": \"enode://8cec1501ea4bb26b6404809e4e3eef53149b490ec7d378085dd7a24b9e6d1211@45.33.89.236:10086\"},{\"Pubkey\": \"0x0398586c5d012e677ef9eb74785de6aa69e38c13e17e45071ec91c9e32bdc64848\",\"Node\": \"enode://5a764d497a55e42ea33a833770f63e2ba8f6200da824db32ef2bcc937af88c44@176.58.112.109:10086\"},{\"Pubkey\": \"0x03ce8088cd8ab8fa7cfd7ecd71015bf525d9b071b0090889e399bc021e10496770\",\"Node\": \"enode://956f3f79214db8065298e8447f8aa14ca611c02bb886a1f1e0be92eb74ad8984@172.104.179.82:10086\"},{\"Pubkey\": \"0x022d06d1131b740ee6324d35c4096ce5dc54a90c0b962b795fd46fa371bc734e82\",\"Node\": \"enode://8ecedb07fcb27642c1560e000c79fa80a4e185001a40030fb50e698ff008a98d@139.162.188.229:10086\"},{\"Pubkey\": \"0x0306c3c718e9f9fad21a63c27e1527c599ec54611e17d6b90fa9a81dc7b6588624\",\"Node\": \"enode://f2055090bb60b42c91272ce00f671c0048c46e065d489910341c101d4bb94af3@23.239.16.64:10086\"},{\"Pubkey\": \"0x0373d68e621f5063875c9b4647cb7b990d96f59fb3769cbb892e462975937acb48\",\"Node\": \"enode://772c23d1de1ac72451c17da61889b5c211127393916f93a8671c202aad2ef9df@45.56.124.251:10086\"},{\"Pubkey\": \"0x02b6d93dddb0cb702266f55b7982067d4b6f5587af0273b660db8ac590d405733f\",\"Node\": \"enode://d8eb92d787384442e40476fa1d4c564facca3667298566b3950554ce4cc0036f@172.104.106.112:10086\"},{\"Pubkey\": \"0x02da50f6372568a6f8dc9d05a65659dc7d5db7db5c1c9bdc4369c70ceb19200b28\",\"Node\": \"enode://6890bf120d755a498a301e9cbbcbabba0c4c6628133c35e5624eb85b5e1f09ae@172.105.175.246:10086\"},{\"Pubkey\": \"0x03d94c1a566bf73783dbe21bfb4fb7b3e673de16a94f5b616c32bd4d44a2bf30c9\",\"Node\": \"enode://f0f207d9199a246871759b29424358bc23bb9d2e3730dc3b151c9f9abdddd4be@172.105.206.6:10086\"},{\"Pubkey\": \"0x03c1586860d81359a51655dfe03663beb4db51ec5373f7b39a525561498dfe31b6\",\"Node\": \"enode://5969d83137afd2ea7ed06419009ea52f86213094b46c6444468880f45659fedd@139.162.30.122:10086\"}]}"

	if args[3] == "testnet" {
		staticConfigs = []string{"enode://548c58daf6dc65d463c155027fce3a909d555683543d1dca34cff1d68868c54f@39.100.111.74:44444", "enode://385c49f05a235115515d5581485be6cd66bbcaf2dbace93d641b5e4c87c20255@39.98.39.224:44444", "enode://9296c4f6e4ceaaea24d0416f49bf7624e920d1f71f7a51877a5d0ed156e35ac5@39.99.44.60:44444"}
		genesisConfig = "{\"Preminer\": [{ \"Addr\": \"0x7d17376a5a611c768970f7ce99fbe309450bff6f\", \"Value\": 10000000000000000000000000000 } ], \"Miners\": [ { \"Pubkey\": \"0x0328378210fd26ac195c4880b5cf8a68e5477d5f2f409e4526ed6b49681091a391\", \"Node\": \"enode://548c58daf6dc65d463c155027fce3a909d555683543d1dca34cff1d68868c54f@39.100.111.74:44444\" }, { \"Pubkey\": \"0x03efe1cad6eb9e161a9d4809eb0d40e9d9392d70e877cc2b41cd7a7526628ee007\", \"Node\": \"enode://385c49f05a235115515d5581485be6cd66bbcaf2dbace93d641b5e4c87c20255@39.98.39.224:44444\" }, { \"Pubkey\": \"0x031afda919527b8c55997e8a2c2cdf33fc025708a73085b4f7f5c500a6c68ddb08\", \"Node\": \"enode://9296c4f6e4ceaaea24d0416f49bf7624e920d1f71f7a51877a5d0ed156e35ac5@39.99.44.60:44444\"}]}"
	}


	err := json.Unmarshal([]byte(genesisConfig), &genesis)
	if err != nil {
		fmt.Println("unmarshal genesis info err")
		return
	}

	trustNodes := []*enode.Node{}
	for i := 0; i < len(staticConfigs); i++ {
		node, err := enode.ParseV4(staticConfigs[i])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		trustNodes = append(trustNodes, node)
	}

	node := getAccount(args[1])
	standbyKey := node.PrivateKey

	logConfig := log.LogConfig{}
	logConfig.LogLevel = 4

	rpcConfig := rpc.RpcConfig{}
	rpcConfig.IPCEnabled = true
	rpcConfig.HTTPEnabled = true

	p2pConfig := p2pTypes.P2pConfig{}
	p2pConfig.MaxPeers = 25
	p2pConfig.NoDiscovery = false
	p2pConfig.DiscoveryV5 = true
	p2pConfig.Name = "drepnode"
	p2pConfig.StaticNodes = trustNodes

	consensusConfig := &cservice.ConsensusConfig{}
	consensusConfig.ConsensusMode = "bft"
	consensusConfig.Bft = &bft.BftConfig{
		MyPk:           nil,
		StartMiner:     true,
		BlockInterval:  15,
		ProducerNum:    21,
		ChangeInterval: 100,
	}

	chainConfig := chain.ChainConfig{}
	chainConfig.ChainId = 0
	chainConfig.GenesisAddr = params.HoleAddress

	chainIndexerConfig := chainIndexerTypes.ChainIndexerConfig{}
	chainIndexerConfig.Enable = true
	chainIndexerConfig.SectionSize = 4096
	chainIndexerConfig.ConfirmsReq = 256
	chainIndexerConfig.Throttling = 100 * time.Millisecond

	filterConfig := filterTypes.FilterConfig{}
	filterConfig.Enable = true

	consensusConfig.Bft.MyPk = (*secp256k1.PublicKey)(&standbyKey.PublicKey)

	p2pConfig.ListenAddr = "0.0.0.0:10086"
	chainConfig.RemotePort = 10087

	os.MkdirAll(userDir, os.ModeDir|os.ModePerm)
	keyStorePath := path2.Join(userDir, "keystore")

	store := accountComponent.NewFileStore(keyStorePath)
	store.StoreKey(node, password)

	walletConfig := accountTypes.Config{}
	walletConfig.Enable = true

	cfgPath := path2.Join(userDir, "config.json")
	fs, _ := os.Create(cfgPath)
	offset := int64(0)
	fs.WriteAt([]byte("{\n"), offset)
	offset = int64(2)

	offset = writePhase(fs, "log", logConfig, offset)
	offset = writePhase(fs, "rpc", rpcConfig, offset)
	offset = writePhase(fs, "consensus", consensusConfig, offset)
	offset = writePhase(fs, "p2p", p2pConfig, offset)
	offset = writePhase(fs, "chain", chainConfig, offset)
	offset = writePhase(fs, "accounts", walletConfig, offset)
	offset = writePhase(fs, "chain_indexer", chainIndexerConfig, offset)
	offset = writePhase(fs, "filter", filterConfig, offset)
	offset = writePhase(fs, "genesis",genesis, offset)
	err = fs.Truncate(offset - 2)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, err = fs.WriteAt([]byte("\n}"), offset-2)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	return
}

func writePhase(fs *os.File, name string, config interface{}, offset int64) int64 {
	bytes, _ := json.MarshalIndent(config, "	", "      ")
	bytes = append([]byte("	\""+name+"\" : "), bytes...)
	fs.WriteAt(bytes, offset)
	offset += int64(len(bytes))

	fs.WriteAt([]byte(",\n"), offset)
	offset += 2
	return offset
}

func getAccount(name string) *types.Node {
	node := RandomNode([]byte(name))
	return node
}

func RandomNode(seed []byte) *types.Node {
	var (
		prvKey    *secp256k1.PrivateKey
		chainCode []byte
	)

	prvKey, _ = crypto.GenerateKey(rand.Reader)
	chainCode = append(seed, []byte(types.DrepMark)...)
	chainCode = common.HmAC(chainCode, prvKey.PubKey().Serialize())

	addr := crypto.PubkeyToAddress(prvKey.PubKey())
	return &types.Node{
		PrivateKey: prvKey,
		Address:    &addr,
		ChainId:    0,
		ChainCode:  chainCode,
	}
}