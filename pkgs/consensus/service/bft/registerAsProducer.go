package bft

import (
	"errors"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/chain/store"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
)

var (
	MinerPrefix = []byte("miner")
	_ = (chain.ITransactionSelector)((*RegisterAsProducerTransactionSelector)(nil))
	_ = (chain.ITransactionValidator)((*RegisterAsProducerTransactionExecutor)(nil))
)

// ***********DEPLOY**************//
type RegisterAsProducerTransactionSelector struct{}

func (registerAsProducerTransactionSelector *RegisterAsProducerTransactionSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.RegisterProducer
}

type RegisterAsProducerTransactionExecutor struct {
}

func (registerAsProducerTransactionExecutor *RegisterAsProducerTransactionExecutor) ExecuteTransaction(context *chain.ExecuteTransactionContext) ([]byte, bool, []*types.Log, error) {
		from := context.From()
		data :=context.Data()
		newProducer := &Producer{}
		err := binary.Unmarshal(data, newProducer)
		if err == nil {
			return nil, false, nil ,err
		}
		if *from == crypto.PubkeyToAddress(newProducer.Pubkey) {
			return nil, false, nil, errors.New("only register himself")
		}
		op := ConsensusOp{context.TrieStore()}
		oldProducers, err := op.GetProducer()
		if err == nil {
			return nil, false, nil ,err
		}

		exit := false
		for _, oldProducer := range oldProducers {
			if oldProducer.Address() == newProducer.Address() {
				exit = true
				oldProducer.Pubkey = newProducer.Pubkey
				oldProducer.IP = newProducer.IP
			}
		}
		if !exit {
			oldProducers = append(oldProducers, newProducer)
		}

		err = op.SaveProducer(oldProducers)
		if err == nil {
			return nil, false, nil ,err
		}
		return nil,true,nil,nil
}

type ConsensusOp struct {
	store.StoreInterface
}

func (consensusOp *ConsensusOp) SaveProducer(p []*Producer) error {
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

func (consensusOp *ConsensusOp) GetProducer() ([]*Producer, error) {
	producers := []*Producer{}
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