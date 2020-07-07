package bft

import (
	"encoding/json"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/types"
)

type MinerGenesisProcessor struct {
}

func NewMinerGenesisProcessor() *MinerGenesisProcessor {
	return &MinerGenesisProcessor{}
}

func (minerGenesisProcessor *MinerGenesisProcessor) Genesis(context *chain.GenesisContext) error {

	val, ok := context.Config()["Miners"]
	if ok {
		miners := []*types.Producer{}
		bytes, _ := val.MarshalJSON()
		err := json.Unmarshal(bytes, &miners) //parserjson
		if err != nil {
			return err
		}

		op := ConsensusOp{context.Store()}
		err = op.SaveProducer(miners) // binary serilize and save to trie
		if err != nil {
			return err
		}
	} else {
		op := ConsensusOp{context.Store()}
		err := op.SaveProducer(chain.DefaultGenesisConfig.Miners) // binary serilize and save to trie
		if err != nil {
			return err
		}
	}
	return nil

}
