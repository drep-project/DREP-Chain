package bft

import (
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/params"
	"math"
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
func (calculator *RewardCalculator) AccumulateRewards(height uint64) error {
	reward := big.NewInt(params.Rewards)
	reward.Mul(reward, new(big.Int).SetUint64(params.Coin))

	rate := int64(height / (4 * params.BlockCountOfEveryYear)) //Number of new blocks in 4 years
	rate = int64(math.Exp2(float64(rate)))
	reward.Div(reward, new(big.Int).SetInt64(rate))

	//Eighty percent for themselves and twenty percent for their supporters
	var selfProportion int64 = 80
	leaderReward := new(big.Int)
	leaderReward = leaderReward.Mul(reward, new(big.Int).SetInt64(selfProportion))
	leaderReward = leaderReward.Div(leaderReward, new(big.Int).SetInt64(100))
	leaderReward.Add(leaderReward, calculator.totalGasBalance)
	leaderAddr := calculator.producers[calculator.sig.Leader].Address()
	err := calculator.trieStore.AddBalance(&leaderAddr, calculator.height, leaderReward)
	if err != nil {
		return err
	}

	//Reward supporters in proportion
	//Distribute Bonus
	otherReward := new(big.Int)
	otherReward = otherReward.Mul(reward, new(big.Int).SetInt64(100-selfProportion))
	otherReward = otherReward.Div(otherReward, new(big.Int).SetInt64(100))

	//fmt.Println(leaderReward.String(), otherReward.String())
	total := new(big.Int)
	supporters := calculator.trieStore.GetCreditDetails(&leaderAddr)

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

	//num := calculator.sig.Num() - 1
	//for index, isCommit := range calculator.sig.Bitmap {
	//	if isCommit == 1 {
	//		addr := calculator.producers[index].Address()
	//		if addr != leaderAddr {
	//			r.Div(reward, new(big.Int).SetInt64(int64(num*2)))
	//			err = calculator.trieStore.AddBalance(&addr, calculator.height, r)
	//			if err != nil {
	//				return err
	//			}
	//		}
	//	}
	//}
	return nil
}
