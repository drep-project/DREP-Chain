package bft

import (
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/params"
	"math/big"
)

type IRewardCalculator interface {
	AccumulateRewards(sig *MultiSignature, Producers ProducerSet, totalGasBalance *big.Int, height uint64)
}

type RewardCalculator struct {
	trieStore       store.StoreInterface
	height          uint64
	sig             *MultiSignature
	producers       ProducerSet
	totalGasBalance *big.Int
}

func NewRewardCalculator(trieStore store.StoreInterface, sig *MultiSignature, producers ProducerSet, totalGasBalance *big.Int, height uint64) *RewardCalculator {
	return &RewardCalculator{
		trieStore:       trieStore,
		sig:             sig,
		producers:       producers,
		totalGasBalance: totalGasBalance,
		height:          height,
	}
}

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func (calculator *RewardCalculator) AccumulateRewards() error {
	reward := big.NewInt(params.Rewards)
	reward.Mul(reward, new(big.Int).SetUint64(params.Coin))

	r := new(big.Int)
	r = r.Div(reward, new(big.Int).SetInt64(2))
	r.Add(r, calculator.totalGasBalance)
	leaderAddr := calculator.producers[calculator.sig.Leader].Address()
	err := calculator.trieStore.AddBalance(&leaderAddr, calculator.height, r)
	if err != nil {
		return err
	}

	num := calculator.sig.Num() - 1
	for index, isCommit := range calculator.sig.Bitmap {
		if isCommit == 1 {
			addr := calculator.producers[index].Address()
			if addr != leaderAddr {
				r.Div(reward, new(big.Int).SetInt64(int64(num*2)))
				err = calculator.trieStore.AddBalance(&addr, calculator.height, r)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
