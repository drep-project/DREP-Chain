package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/drep-project/DREP-Chain/common/hexutil"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/pkgs/accounts/service"
	"github.com/drep-project/DREP-Chain/pkgs/trace"
	"math/big"
	"net/url"
	"os"
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
	if len(args) <= 1 {
		return fmt.Errorf("arg num too small")
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
			fmt.Println("cannot deal method:", args[0], "need register it")
		}
		wg.Done()
	}()

	wg.Wait()
	return nil
}

func getBalance(args cli.Args, client *rpc.Client, ctx context.Context) {
	balance := ""
	if err := client.CallContext(ctx, &balance, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
	}

	fmt.Println(balance)
}

func createCode(args cli.Args, client *rpc.Client, ctx context.Context) {
	r := ""
	if err := client.CallContext(ctx, &r, args[0], args[1], args[2], args[3], args[4]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(r)
}

func executeContract(args cli.Args, client *rpc.Client, ctx context.Context) {
	if len(args) != 6 {
		fmt.Println("param num:", len(args), "must equal 6")
		return
	}

	r := ""
	if err := client.CallContext(ctx, &r, args[0], args[1], args[2], args[3], args[4], args[5]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(r)
}

func getPoolTransactions(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]types.Transactions, 0, 2)
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

func init() {
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

func candidateCredit(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5], args[6]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func sendRawTransaction(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
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
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getPoolMiniPendingNonce(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp uint64
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getTxInPool(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Transaction
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getBlock(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Block
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
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getReputation(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := new(big.Int)
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getTransactionByBlockHeightAndIndex(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Transaction

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
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getAddressByAlias(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getReceipt(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Receipt
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}

func getLogs(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]*types.Log, 0)
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getCancelCreditDetailByTXHash(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]*types.CancelCreditDetail, 0)
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getByteCode(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp hexutil.Bytes
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp.String())
}
func getCreditDetails(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func getCancelCreditDetails(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
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
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
	}
	fmt.Println("success")
}
func removePeer(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp interface{}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func setLevel(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp interface{}
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func setVmodule(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func getRawTransaction(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func getTransaction(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp types.Transaction
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func decodeTrasnaction(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp trace.RpcTransaction
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getSendTransactionByAddr(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]*trace.RpcTransaction, 0)
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
func getReceiveTransactionByAddr(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]*trace.RpcTransaction, 0)
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func rebuild(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
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
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func createWallet(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("success")
}
func lockAccount(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("lock success")
}
func unlockAccount(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println("unlock success")
}
func openWallet(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp error
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
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5], args[6]); err != nil {
		fmt.Println("return err :", err)
	}
	fmt.Println(resp)
}
func transferWithNonce(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	oldNonce, _ := strconv.Atoi(args[7])
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5], args[6], oldNonce); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func setAlias(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string //error
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func voteCredit(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string //(string, error)
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func cancelVoteCredit(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string //error
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func cancelCandidateCredit(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3], args[4], args[5]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func readContract(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp common.Bytes //common.Bytes, error
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], args[3]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func estimateGas(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp uint64 //(uint64, error) {
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2], common.Bytes(args[3]), args[4]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}

func dumpPrivkey(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp secp256k1.PrivateKey //*secp256k1.PrivateKey, error
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	fmt.Println("0x"+common.Bytes2Hex(resp.Serialize()))
}
func dumpPubkey(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp secp256k1.PublicKey //*secp256k1.PublicKey, error
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}

	fmt.Println("0x" + common.Bytes2Hex(resp.Serialize()))
}
func sign(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp common.Bytes //(common.Bytes, error)
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func generateAddresses(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp service.RpcAddresses //(*RpcAddresses, error)
	if err := client.CallContext(ctx, &resp, args[0], args[1]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func importKeyStore(args cli.Args, client *rpc.Client, ctx context.Context) {
	resp := make([]string, 0) //([]string, error)
	if err := client.CallContext(ctx, &resp, args[0], args[1], args[2]); err != nil {
		fmt.Println("return err :", err)
		return
	}
	fmt.Println(resp)
}
func importPrivkey(args cli.Args, client *rpc.Client, ctx context.Context) {
	var resp string // (string, error)
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
