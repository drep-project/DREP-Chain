package pool

import (
    "BlockChainTest/util/objectemitter"
    "time"
    "BlockChainTest/bean"
)

var (
    txPool *objectemitter.ObjectEmitter
)

func StartTxPool() {
    txPool = objectemitter.New(1000, 10 * time.Second, func(txs []interface{}) {

    })
    txPool.Start()
}
func PushTx(txs []*bean.Transaction) {
    for _, tx := range txs {
        txPool.Push(tx)
    }
}
