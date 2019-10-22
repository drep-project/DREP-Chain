package chain

import (
	"encoding/json"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/chain/store"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
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

	val, ok = context.Config()["miners"]
	if ok {
		miners := []types.CandidateData{}
		bytes, _ := val.MarshalJSON()
		err := json.Unmarshal(bytes, &miners)
		if err != nil {
			return err
		}

		addrs := []crypto.CommonAddress{}
		for _, miner := range miners {
			minerBytes, err := json.Marshal(miner)
			if err != nil {
				return err
			}
			addr := crypto.PubkeyToAddress(miner.Pubkey)

			context.Store().CandidateCredit(&addr, new(big.Int).SetUint64(10), minerBytes, 0)
			context.store.AddCandidateAddr(&addr)

			addrs = append(addrs, addr)
		}

		addrsBytes, err := binary.Marshal(addrs)
		if err != nil {
			return err
		}

		err = context.Store().Put([]byte(store.CandidateAddrs), addrsBytes)
		if err != nil {
			return err
		}
	}
	return nil
}
