package bft

import (
	"errors"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
)

var (
	MinerPrefix = []byte("miner")
	_           = (chain.ITransactionSelector)((*RegisterAsProducerTransactionSelector)(nil))
	_           = (chain.ITransactionValidator)((*RegisterAsProducerTransactionExecutor)(nil))
)

// ***********DEPLOY**************//
type RegisterAsProducerTransactionSelector struct{}

func (registerAsProducerTransactionSelector *RegisterAsProducerTransactionSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.RegisterProducer
}

type RegisterAsProducerTransactionExecutor struct {
}

func (registerAsProducerTransactionExecutor *RegisterAsProducerTransactionExecutor) ExecuteTransaction(context *chain.ExecuteTransactionContext) *types.ExecuteTransactionResult {
	etr := &types.ExecuteTransactionResult{}
	from := context.From()
	data := context.Data()
	newProducer := &types.Producer{}
	err := binary.Unmarshal(data, newProducer)
	if err == nil {
		etr.Txerror = err
		return etr
	}
	if *from == crypto.PubkeyToAddress(newProducer.Pubkey) {
		etr.Txerror = errors.New("only register himself")
		return etr
	}
	op := ConsensusOp{context.TrieStore()}
	oldProducers, err := op.GetProducer()
	if err == nil {
		etr.Txerror = err
		return etr
	}

	exit := false
	for _, oldProducer := range oldProducers {
		if oldProducer.Address() == newProducer.Address() {
			exit = true
			oldProducer.Pubkey = newProducer.Pubkey
			oldProducer.Node = newProducer.Node
		}
	}
	if !exit {
		oldProducers = append(oldProducers, newProducer)
	}

	err = op.SaveProducer(oldProducers)
	if err == nil {
		etr.Txerror = err
		return etr
	}
	return etr
}

type ConsensusOp struct {
	store.StoreInterface
}

func (consensusOp *ConsensusOp) SaveProducer(p []*types.Producer) error {
	b, err := binary.Marshal(p)
	if err != nil {
		return err
	}

	err = consensusOp.Put(MinerPrefix, b)
	if err != nil {
		return err
	}
	return nil
}

func (consensusOp *ConsensusOp) GetProducer() ([]*types.Producer, error) {
	producers := []*types.Producer{}
	bytes, err := consensusOp.Get(MinerPrefix)
	if err != nil {
		return nil, err
	}
	err = binary.Unmarshal(bytes, &producers)
	if err != nil {
		return nil, err
	}
	return producers, nil
}
