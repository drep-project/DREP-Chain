package chain

import (
	"github.com/drep-project/drep-chain/params"
	"github.com/drep-project/drep-chain/types"
	"math/big"
)

/**********************alias********************/

type AliasTxSelector struct {
}

func (liasTxSelector *AliasTxSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.SetAliasType
}

var (
	_ = (ITransactionSelector)((*AliasTxSelector)(nil))
	_ = (ITransactionValidator)((*AliasTransactionProcessor)(nil))
)

type AliasTransactionProcessor struct {
}

//5 160000 640
//6 80000 320
//7 40000 160
//8 20000 80
//9 10000 40
//10 5000 20
//11 2500 10
func (aliasTransactionProcessor *AliasTransactionProcessor) ExecuteTransaction(context *ExecuteTransactionContext) ([]byte, bool, []*types.Log, error) {
	from := context.From()
	store := context.TrieStore()
	tx := context.Tx()
	alias := tx.GetData()
	if err := CheckAlias(alias); err != nil {
		return nil, false, nil, err
	}
	err := store.AliasSet(from, string(alias))
	if err != nil {
		return nil, false, nil, err
	}
	err = context.UseGas(params.AliasGas * uint64(len(alias)))
	if err != nil {
		return nil, false, nil, err
	}
	// extra price
	type LenPriceCacler struct {
		LenMatch func() bool
		Fee      func() *big.Int
	}

	calcers := []*LenPriceCacler{
		{
			LenMatch: func() bool {
				return len(alias) == 5
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(160000)
			},
		},
		{
			LenMatch: func() bool {
				return len(alias) == 6
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(80000)
			},
		},
		{
			LenMatch: func() bool {
				return len(alias) == 7
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(40000)
			},
		},
		{
			LenMatch: func() bool {
				return len(alias) == 8
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(20000)
			},
		},
		{
			LenMatch: func() bool {
				return len(alias) == 9
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(10000)
			},
		},
		{
			LenMatch: func() bool {
				return len(alias) == 10
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(5000)
			},
		},
		{
			LenMatch: func() bool {
				return len(alias) == 11
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(2500)

			},
		},
		{
			LenMatch: func() bool {
				return len(alias) > 11
			},
			Fee: func() *big.Int {
				return big.NewInt(0)
			},
		},
	}
	var drepFee *big.Int
	for _, calcer := range calcers {
		if calcer.LenMatch() {
			drepFee = calcer.Fee()
			break
		}
	}

	//minus alias fee from from account
	originBalance := store.GetBalance(from)
	leftBalance := originBalance.Sub(originBalance, drepFee)
	if leftBalance.Sign() < 0 {
		return nil, false, nil, ErrBalance
	}
	err = store.PutBalance(from, leftBalance)
	if err != nil {
		return nil, false, nil, err
	}
	// put alias fee to hole address
	zeroAddressBalance := store.GetBalance(&params.HoleAddress)
	zeroAddressBalance = zeroAddressBalance.Add(zeroAddressBalance, drepFee)
	err = store.PutBalance(&params.HoleAddress, zeroAddressBalance)
	if err != nil {
		return nil, false, nil, err
	}
	err = store.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		return nil, false, nil, err
	}

	return nil, true, nil, err
}

func CheckAlias(alias []byte) error {

	if len(alias) < 5 {
		return ErrTooShortAlias
	}
	if len(alias) > 20 {
		return ErrTooLongAlias
	}

	runes := []rune(string(alias))
	for i := 0; i < len(runes); i++ {
		//number  48-57
		if 48 <= runes[i] && runes[i] <= 57 {
			continue
		}
		//upcase
		if 65 <= runes[i] && runes[i] <= 90 {
			continue
		}
		//lowcaser
		if 97 <= runes[i] && runes[i] <= 122 {
			continue
		}
		return ErrUnsupportAliasChar
	}
	return nil
}
