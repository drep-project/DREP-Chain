package bft

import (
	"encoding/json"
	"github.com/drep-project/drep-chain/chain"
)

type MinerGenesisProcessor struct {
}

func NewMinerGenesisProcessor() *MinerGenesisProcessor {
	return &MinerGenesisProcessor{}
}

func (minerGenesisProcessor *MinerGenesisProcessor) Genesis(context *chain.GenesisContext) error {

	val, ok := context.Config()["miner"]
	if ok {
		miners := []*Producer{}
		bytes, _ := val.MarshalJSON()
		err := json.Unmarshal(bytes, &miners)  //parserjson
		if err != nil {
			return err
		}

		op := ConsensusOp{context.Store()}
		err = op.SaveProducer(miners)		// binary serilize and save to trie
		if err != nil {
			return err
		}
	}
	return nil

}

