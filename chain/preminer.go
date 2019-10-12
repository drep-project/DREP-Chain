package chain

import (
	"encoding/json"
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
)

type Preminer struct {
	Addr  crypto.CommonAddress
	Value big.Int
}

type PreminerGenesisProcessor struct {
}

func NewPreminerGenesisProcessor() *PreminerGenesisProcessor {
	return &PreminerGenesisProcessor{}
}

func (NewPreminerGenesisProcessor *PreminerGenesisProcessor) Genesis(context *GenesisContext) error {
	val, ok := context.Config()["preminer"]
	if ok {
		preminers := []Preminer{}
		bytes, _ := val.MarshalJSON()
		err := json.Unmarshal(bytes, &preminers)
		if err != nil {
			return err
		}

		for _, preminer := range preminers {
			err = context.Store().PutBalance(&preminer.Addr, 0, &preminer.Value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
