package blockmgr

import (
	"fmt"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
	"math/big"
	"time"
)

type TemplateBlockValidator struct {
	chain chain.ChainServiceInterface
}

func NewTemplateBlockValidator(chain chain.ChainServiceInterface) *TemplateBlockValidator {
	return &TemplateBlockValidator{chain}
}

func (chainBlockValidator *TemplateBlockValidator) VerifyHeader(header, parent *types.BlockHeader) error {
	return nil
}

func (chainBlockValidator *TemplateBlockValidator) VerifyBody(block *types.Block) error {
	return nil
}

func (chainBlockValidator *TemplateBlockValidator) ExecuteBlock(context *chain.BlockExecuteContext) error {
	context.Receipts = make([]*types.Receipt, context.Block.Data.TxCount)
	context.Logs = make([]*types.Log, 0)
	if len(context.Block.Data.TxList) < 0 {
		return nil
	}

	finalTxs := make([]*types.Transaction, 0, len(context.Block.Data.TxList))
	finalReceipts := make([]*types.Receipt, 0, len(context.Block.Data.TxList))
	//time control
	stopchanel := make(chan struct{}, 1)
	tm := time.AfterFunc(time.Second*5, func() {
		stopchanel <- struct{}{}
	})
	defer func() {
		context.Block.Data.TxList = finalTxs
		context.Block.Data.TxCount = uint64(len(finalTxs))
		context.Receipts = finalReceipts
		context.Block.Header.GasUsed = *context.GasUsed
		context.Block.Header.TxRoot = chainBlockValidator.chain.DeriveMerkleRoot(finalTxs)
		context.Block.Header.ReceiptRoot = chainBlockValidator.chain.DeriveReceiptRoot(finalReceipts)
	}()
SELECT_TX:
	for _, t := range context.Block.Data.TxList {
		snap := context.TrieStore.CopyState()
		backGp := *context.Gp
		select {
		case <-stopchanel:
			break SELECT_TX
		default:
			receipt, gasUsed, err := chainBlockValidator.RouteTransaction(context, context.Gp, t)
			if err == nil {
				finalTxs = append(finalTxs, t)
				finalReceipts = append(finalReceipts, receipt)
				gasUsedBig := new(big.Int).SetUint64(gasUsed)
				context.AddGasUsed(gasUsedBig)
				gasFee := new(big.Int).Mul(gasUsedBig, t.GasPrice())
				context.AddGasFee(gasFee)
			} else if err == chain.ErrOutOfGas {
				// return while out of gas
				context.TrieStore.RevertState(snap)
				context.Gp = &backGp
				return nil
			} else {
				from,_ := t.From()
				log.WithField("err",err).WithField("from", from.String()).WithField("tx nonce",t.Nonce()).Info("route tx")
				//skip wrong tx
				context.TrieStore.RevertState(snap)
				context.Gp = &backGp
				continue
			}
		}
	}
	tm.Stop()
	return nil
}

func (chainBlockValidator *TemplateBlockValidator) RouteTransaction(context *chain.BlockExecuteContext, gasPool *chain.GasPool, tx *types.Transaction) (*types.Receipt, uint64, error) {
	//init transaction tx
	from, err := tx.From()
	if err != nil {
		return nil, 0, err
	}

	txContext := chain.NewExecuteTransactionContext(context, context.TrieStore, gasPool, from, tx)
	if err := txContext.PreCheck(); err != nil {
		return nil, 0, err
	}

	// Pay intrinsic gastx
	gas, err := tx.IntrinsicGas()
	if err != nil {
		return nil, 0, err
	}

	if err = txContext.UseGas(gas); err != nil {
		return nil, 0, err
	}
	exit := false
	for selector, txValidator := range chainBlockValidator.chain.TransactionValidators() {
		if selector.Select(tx) {
			exit = true
			_, failed, logs, err := txValidator.ExecuteTransaction(txContext)
			if err != nil {
				return nil, 0, err
			}
			err = txContext.RefundCoin()
			if err != nil {
				return nil, 0, err
			}
			// Create a new receipt for the transaction, storing the intermediate root and gasRemained used by the tx
			// based on the eip phase, we're passing whether the root touch-delete accounts.
			receipt := types.NewReceipt(crypto.ZeroHash[:], failed, txContext.GasUsed())
			receipt.TxHash = *tx.TxHash()
			receipt.GasUsed = txContext.GasUsed()
			// if the transaction created a contract, store the creation address in the receipt.
			if tx.To() == nil || tx.To().IsEmpty() {
				receipt.ContractAddress = crypto.CreateAddress(*from, tx.Nonce())
				fmt.Println(receipt.ContractAddress)
			}
			// Set the receipt logs and create a bloom for filtering
			receipt.Logs = logs
			receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
			//receipt.BlockHash = *header.Hash()
			receipt.BlockNumber = context.Block.Header.Height
			return receipt, txContext.GasUsed(), nil
		}
	}
	if !exit {
		return nil, 0, chain.ErrUnsupportTxType
	}
	panic("never come here")
}
