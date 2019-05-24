package blockmgr

import (
	"github.com/drep-project/binary"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"math/big"
)

type BlockMgrApi struct {
	blockMgr *BlockMgr
	dbService    *database.DatabaseService
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

func (blockMgrApi *BlockMgrApi)  GasPrice() (*big.Int, error) {
	return blockMgrApi.blockMgr.gpo.SuggestPrice()
}

func (blockMgrApi *BlockMgrApi)  GetPoolTransactions(addr *crypto.CommonAddress) []chainTypes.Transactions {
	return blockMgrApi.blockMgr.GetPoolTransactions(addr)
}

func (blockMgrApi *BlockMgrApi)  GetPoolMiniPendingNonce(addr *crypto.CommonAddress) uint64 {
	return blockMgrApi.blockMgr.GetPoolMiniPendingNonce(addr)
}


func (blockMgrApi *BlockMgrApi) GenerateTransferTransaction(to  *crypto.CommonAddress, nonce uint64, amount, price, limit common.Big) chainTypes.Transaction {
	return blockMgrApi.blockMgr.GenerateTransferTransaction(to, nonce, amount, price, limit)
}
