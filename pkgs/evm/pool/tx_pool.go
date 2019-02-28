package pool

import (
    "github.com/drep-project/drep-chain/common/objectemitter"
    "time"
    chainTypes "github.com/drep-project/drep-chain/chain/types"
)

var (
    txPool *objectemitter.ObjectEmitter
)

func StartTxPool() {
    txPool = objectemitter.New(1000, 10 * time.Second, func(txs []interface{}) {

    })
    txPool.Start()
}
func PushTx(txs []*chainTypes.Transaction) {
    for _, tx := range txs {
        txPool.Push(tx)
    }
}
