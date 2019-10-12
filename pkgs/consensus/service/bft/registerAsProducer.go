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
		p := &Producer{}
		err := binary.Unmarshal(data, p)
		if err == nil {
			return nil, false, nil ,err
		}
		if *from == crypto.PubkeyToAddress(p.Pubkey) {
			return nil, false, nil, errors.New("only register himself")
		}
		op := ConsensusOp{context.TrieStore()}
		producers, err := op.GetProducer()
		if err == nil {
			return nil, false, nil ,err
		}
		producers[*from] = *p
		err = op.SaveProducer(producers)
		if err == nil {
			return nil, false, nil ,err
		}
		return nil,true,nil,nil
}

type ConsensusOp struct {
	store.StoreInterface
}

func (consensusOp *ConsensusOp) SaveProducer(p map[crypto.CommonAddress]Producer) error {
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

func (consensusOp *ConsensusOp) GetProducer() (map[crypto.CommonAddress]Producer, error) {
	producers := make(map[crypto.CommonAddress]Producer)
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