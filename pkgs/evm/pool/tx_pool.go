package pool

import (
	"github.com/drep-project/drep-chain/types"
	"github.com/drep-project/drep-chain/common/objectemitter"
	"time"
)

var (
	txPool *objectemitter.ObjectEmitter
)

func StartTxPool() {
	txPool = objectemitter.New(1000, 10*time.Second, func(txs []interface{}) {

	})
	txPool.Start()
}
func PushTx(txs []*types.Transaction) {
	for _, tx := range txs {
		txPool.Push(tx)
	}
}
