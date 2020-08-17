package bft

import (
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/types"
	"math"
	"math/big"
)

type IRewardCalculator interface {
	AccumulateRewards(sig *MultiSignature, Producers types.ProducerSet, totalGasBalance *big.Int, height uint64)
}

type RewardCalculator struct {
	trieStore       store.StoreInterface
	height          uint64
	sig             *MultiSignature
	producers       types.ProducerSet
	totalGasBalance *big.Int
}

func NewRewardCalculator(trieStore store.StoreInterface, sig *MultiSignature, producers types.ProducerSet, totalGasBalance *big.Int, height uint64) *RewardCalculator {
	return &RewardCalculator{
		trieStore:       trieStore,
		sig:             sig,
		producers:       producers,
		totalGasBalance: totalGasBalance,
		height:          height,
	}
}

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func (calculator *RewardCalculator) AccumulateRewards(height uint64) error {
	reward := big.NewInt(params.Rewards)
	reward.Mul(reward, new(big.Int).SetUint64(params.Coin))

	rate := int64(height / (4 * params.BlockCountOfEveryYear)) //Number of new blocks in 4 years
	rate = int64(math.Exp2(float64(rate)))
	reward.Div(reward, new(big.Int).SetInt64(rate))

	//Eighty percent for themselves and twenty percent for their supporters
	var selfProportion int64 = 80
	leaderAddr := calculator.producers[calculator.sig.Leader].Address()

	//Reward supporters in proportion
	//Distribute Bonus
	otherReward := new(big.Int)
	otherReward = otherReward.Mul(reward, new(big.Int).SetInt64(100-selfProportion))
	otherReward = otherReward.Div(otherReward, new(big.Int).SetInt64(100))

	total := new(big.Int)
	supporters := calculator.trieStore.GetCreditDetails(&leaderAddr)
	delete(supporters, leaderAddr)

	for _, v := range supporters {
		total = total.Add(total, &v)
	}

	for spporterAddr, supportCredit := range supporters {
		bonus := new(big.Int).Set(otherReward)
		bonus = bonus.Mul(bonus, &supportCredit)
		bonus = bonus.Div(bonus, total)

		err := calculator.trieStore.AddBalance(&spporterAddr, calculator.height, bonus)
		if err != nil {
			return err
		}
	}

	if len(supporters) == 0 {
		selfProportion = 100 //没有支持者，自己获得100%的收入
	}

	leaderReward := new(big.Int)
	leaderReward = leaderReward.Mul(reward, new(big.Int).SetInt64(selfProportion))
	leaderReward = leaderReward.Div(leaderReward, new(big.Int).SetInt64(100))
	leaderReward.Add(leaderReward, calculator.totalGasBalance)

	err := calculator.trieStore.AddBalance(&leaderAddr, calculator.height, leaderReward)
	if err != nil {
		return err
	}

	return nil
}
