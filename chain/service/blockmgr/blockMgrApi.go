package blockmgr

import (
	"github.com/drep-project/binary"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"math/big"
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

func (blockMgrApi *BlockMgrApi) SendRawTransaction(txbytes common.Bytes) (string, error) {
	tx := &chainTypes.Transaction{}
	err := binary.Unmarshal(txbytes, tx)
	if err != nil {
		return "", err
	}
	err = blockMgrApi.blockMgr.SendTransaction(tx, true)
	if err != nil {
		return "", err
	}
	blockMgrApi.blockMgr.BroadcastTx(chainTypes.MsgTypeTransaction, tx, true)

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
func (blockMgrApi *BlockMgrApi) GetPoolTransactions(addr *crypto.CommonAddress) []chainTypes.Transactions {
	return blockMgrApi.blockMgr.GetPoolTransactions(addr)
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

func (blockMgrApi *BlockMgrApi) GenerateTransferTransaction(to *crypto.CommonAddress, nonce uint64, amount, price, limit common.Big) chainTypes.Transaction {
	return blockMgrApi.blockMgr.GenerateTransferTransaction(to, nonce, amount, price, limit)
}
