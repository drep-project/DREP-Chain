package chain

import (
	"github.com/drep-project/drep-chain/chain/store"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
	"math/big"
)

type ITransactionValidator interface {
	ExecuteTransaction(context *ExecuteTransactionContext) ([]byte, bool, []*types.Log, error)
}

type ITransactionSelector interface {
	Select(tx *types.Transaction) bool
}

type ExecuteTransactionContext struct {
	blockContext *BlockExecuteContext
	trieStore    store.StoreInterface

	gp          *GasPool
	tx          *types.Transaction
	from        *crypto.CommonAddress
	gasPrice    *big.Int
	value       *big.Int
	data        []byte
	header      *types.BlockHeader
	gasRemained uint64
	initialGas  uint64
}

func NewExecuteTransactionContext(blockContext *BlockExecuteContext, chainstore store.StoreInterface, gasPool *GasPool, from *crypto.CommonAddress, tx *types.Transaction) *ExecuteTransactionContext {
	context := &ExecuteTransactionContext{trieStore: chainstore, gp: gasPool, tx: tx, from: from}
	context.blockContext = blockContext
	context.from = from
	context.gasPrice = tx.GasPrice()
	context.value = tx.Amount()
	context.data = tx.GetData()
	context.header = blockContext.Block.Header
	return context
}
func (context *ExecuteTransactionContext) Header() *types.BlockHeader {
	return context.header
}

func (context *ExecuteTransactionContext) GasRemained() uint64 {
	return context.gasRemained
}

func (context *ExecuteTransactionContext) RefundGas(refundGas uint64) {
	context.gasRemained = context.gasRemained + refundGas
}

func (context *ExecuteTransactionContext) Data() []byte {
	return context.data
}

func (context *ExecuteTransactionContext) Value() *big.Int {
	return context.value
}

func (context *ExecuteTransactionContext) InitialGas() uint64 {
	return context.initialGas
}

func (context *ExecuteTransactionContext) GasPrice() *big.Int {
	return context.gasPrice
}

func (context *ExecuteTransactionContext) From() *crypto.CommonAddress {
	return context.from
}

func (context *ExecuteTransactionContext) Tx() *types.Transaction {
	return context.tx
}

func (context *ExecuteTransactionContext) Gp() *GasPool {
	return context.gp
}

func (context *ExecuteTransactionContext) TrieStore() store.StoreInterface {
	return context.trieStore
}

func (context *ExecuteTransactionContext) RefundCoin() error {
	// Return DREP for remaining gasRemained, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(context.gasRemained), context.gasPrice)
	err := context.trieStore.AddBalance(context.from, context.header.Height, remaining)
	if err != nil {
		return nil
	}
	// Also return remaining gasRemained to the block gasRemained counter so it is
	// available for the next transaction.
	context.gp.AddGas(context.gasRemained)
	return nil
}

// gasRemained returns the amount of gasRemained used up by the state transition.
func (context *ExecuteTransactionContext) GasUsed() uint64 {
	return context.initialGas - context.gasRemained
}

func (context *ExecuteTransactionContext) To() crypto.CommonAddress {
	if context.tx == nil || context.tx.To() == nil || context.tx.To().IsEmpty() /* contract creation */ {
		return crypto.CommonAddress{}
	}
	return *context.tx.To()
}

func (context *ExecuteTransactionContext) UseGas(amount uint64) error {
	if context.gasRemained < amount {
		return ErrOutOfGas
	}
	context.gasRemained -= amount

	return nil
}

func (context *ExecuteTransactionContext) buyGas() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(context.tx.Gas()), context.gasPrice)
	if context.trieStore.GetBalance(context.from, context.header.Height).Cmp(mgval) < 0 {
		return ErrInsufficientBalanceForGas
	}
	if err := context.gp.SubGas(context.tx.Gas()); err != nil {
		return err
	}
	context.gasRemained += context.tx.Gas()

	context.initialGas = context.tx.Gas()
	return context.trieStore.SubBalance(context.from, context.header.Height, mgval)
}

func (context *ExecuteTransactionContext) PreCheck() error {
	// Make sure this transaction's nonce is correct.
	nonce := context.trieStore.GetNonce(context.from)
	if nonce < context.tx.Nonce() {
		log.WithField("db nonce", nonce).WithField("tx nonce", context.tx.Nonce()).WithField("from",context.from.String()).Info("state precheck too hight")
		return ErrNonceTooHigh
	} else if nonce > context.tx.Nonce() {
		log.WithField("db nonce", nonce).WithField("tx nonce", context.tx.Nonce()).WithField("from",context.from.String()).Info("state precheck too low")
		return ErrNonceTooLow
	}
	return context.buyGas()
}
