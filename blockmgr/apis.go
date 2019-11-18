package blockmgr

import (
	"fmt"
	"math/big"

	"github.com/drep-project/binary"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/types"
)

/*
name: 区块
usage: 用于处理区块链偏上层逻辑
prefix:blockMgr
*/
type BlockMgrApi struct {
	blockMgr  *BlockMgr
	dbService *database.DatabaseService
}

/*
 name: sendRawTransaction
 usage: 获取交易池中的交易信息.
 params:
	1. 待查询地址
 return: 交易池中所有交易
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_sendRawTransaction","params":["0x40a287b6d30b05313131317a4120dd8c23c40910d038fa43b2f8932d3681cbe5ee3079b6e9de0bea6e8e6b2a867a561aa26e1cd6b62aa0422a043186b593b784bf80845c3fd5a7fbfe62e61d8564"], "id": 3}' -H "Content-Type:application/json"
 response:
	{"jsonrpc":"2.0","id":1,"result":"0xf30e858667fa63bc57ae395c3f57ede9bb3ad4969d12f4bce51d900fb5931538"}
*/
func (blockMgrApi *BlockMgrApi) SendRawTransaction(txbytes common.Bytes) (string, error) {
	fmt.Println(string(txbytes))
	tx := &types.Transaction{}
	err := binary.Unmarshal(txbytes, tx)
	if err != nil {
		return "", err
	}
	err = blockMgrApi.blockMgr.SendTransaction(tx, true)
	if err != nil {
		return "", err
	}
	blockMgrApi.blockMgr.BroadcastTx(types.MsgTypeTransaction, tx, true)
	return tx.TxHash().String(), err
}

func (blockMgrApi *BlockMgrApi) GasPrice() (*big.Int, error) {
	return blockMgrApi.blockMgr.gpo.SuggestPrice()
}

/*
 name: GetPoolTransactions
 usage: 获取交易池中的交易信息.
 params:
	1. 待查询地址
 return: 交易池中所有交易
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getPoolTransactions","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
*/
func (blockMgrApi *BlockMgrApi) GetPoolTransactions(addr *crypto.CommonAddress) []types.Transactions {
	return blockMgrApi.blockMgr.GetPoolTransactions(addr)
}

/*
 name: GetTransactionCount
 usage: 获取交易池中的交易信息.
 params:
	1. 待查询地址
 return: 交易池中所有交易
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getTransactionCount","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
*/
func (blockMgrApi *BlockMgrApi) GetTransactionCount(addr *crypto.CommonAddress) uint64 {
	return blockMgrApi.blockMgr.GetTransactionCount(addr)
}

/*
 name: GetPoolMiniPendingNonce
 usage: 获取pending队列中，最小的Nonce
 params:
	1. 待查询地址
 return: pending 队列中最小的nonce
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getPoolMiniPendingNonce","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
*/

func (blockMgrApi *BlockMgrApi) GetPoolMiniPendingNonce(addr *crypto.CommonAddress) uint64 {
	return blockMgrApi.blockMgr.GetPoolMiniPendingNonce(addr)
}
