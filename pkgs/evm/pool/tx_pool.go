package pool

import (
	"github.com/drep-project/DREP-Chain/common/objectemitter"
	"github.com/drep-project/DREP-Chain/types"
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
