package bft

import (
	"encoding/json"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/crypto"
)

type MinerGenesisProcessor struct {
}

func NewMinerGenesisProcessor() *MinerGenesisProcessor {
	return &MinerGenesisProcessor{}
}

func (minerGenesisProcessor *MinerGenesisProcessor) Genesis(context *chain.GenesisContext) error {

	val, ok := context.Config()["miner"]
	if ok {
		miners := []Producer{}
		bytes, _ := val.MarshalJSON()
		err := json.Unmarshal(bytes, &miners)
		if err != nil {
			return err
		}

		op := ConsensusOp{context.Store()}
		producers := map[crypto.CommonAddress]Producer{}
		for _, producer := range miners {
			producers[crypto.PubkeyToAddress(producer.Pubkey)] = producer
		}
		err = op.SaveProducer(producers)
		if err != nil {
			return err
		}
	}
	return nil

}

