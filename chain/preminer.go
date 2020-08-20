package chain

import (
	"encoding/json"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
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

func (PreminerGenesisProcessor *PreminerGenesisProcessor) Genesis(context *GenesisContext) error {
	val, ok := context.Config()["Preminer"]
	preminers := []*Preminer{}
	if ok {
		bytes, _ := val.MarshalJSON()
		err := json.Unmarshal(bytes, &preminers)
		if err != nil {
			return err
		}
	} else {
		preminers = append(preminers, DefaultGenesisConfigMainnet.Preminer...)
	}

	for _, preminer := range preminers {
		err := context.Store().PutBalance(&preminer.Addr, 0, &preminer.Value)
		if err != nil {
			return err
		}
	}

	val, ok = context.Config()["Miners"]
	miners := []*types.Producer{}
	if ok {
		bytes, _ := val.MarshalJSON()
		err := json.Unmarshal(bytes, &miners)
		if err != nil {
			return err
		}
	} else {
		miners = append(miners, DefaultGenesisConfigMainnet.Miners...)
	}

	addrs := []crypto.CommonAddress{}
	for _, miner := range miners {
		minerBytes, err := json.Marshal(miner)
		if err != nil {
			return err
		}
		addr := crypto.PubkeyToAddress(miner.Pubkey)

		//pv := new(big.Int).SetUint64(store.RegisterPledgeLimit)
		pv := new(big.Int).SetUint64(0)
		pv = pv.Mul(pv, new(big.Int).SetUint64(params.Coin))
		context.Store().CandidateCredit(&addr, pv, minerBytes, 0)
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

	return nil
}
