package chain

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/types"
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

type AliasTransactionProcessor struct{}

//5 160000 640
//6 80000 320
//7 40000 160
//8 20000 80
//9 10000 40
//10 5000 20
//11 2500 10
func (aliasTransactionProcessor *AliasTransactionProcessor) ExecuteTransaction(context *ExecuteTransactionContext) *types.ExecuteTransactionResult {
	etr := types.ExecuteTransactionResult{}
	from := context.From()
	store := context.TrieStore()
	tx := context.Tx()
	alias := tx.GetData()
	if err := CheckAlias(tx, store, context.header.Height); err != nil {
		etr.Txerror = err
		return &etr
	}
	err := store.AliasSet(from, string(alias))
	if err != nil {
		etr.Txerror = err
		return &etr
	}
	err = context.UseGas(params.AliasGas * uint64(len(alias)))
	if err != nil {
		etr.Txerror = err
		return &etr
	}

	drepFee := getFee(len(alias))

	//minus alias fee from from account
	originBalance := store.GetBalance(from, context.header.Height)
	leftBalance := originBalance.Sub(originBalance, drepFee)
	if leftBalance.Sign() < 0 {
		etr.Txerror = ErrBalance
		return &etr
	}
	err = store.PutBalance(from, context.header.Height, leftBalance)
	if err != nil {
		etr.Txerror = err
		return &etr
	}
	// put alias fee to hole address
	zeroAddressBalance := store.GetBalance(&params.HoleAddress, context.header.Height)
	zeroAddressBalance = zeroAddressBalance.Add(zeroAddressBalance, drepFee)
	err = store.PutBalance(&params.HoleAddress, context.header.Height, zeroAddressBalance)
	if err != nil {
		etr.Txerror = err
		return &etr
	}
	err = store.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		etr.Txerror = err
		return &etr
	}

	return &etr
}

func getFee(len int) *big.Int {
	// extra price
	type LenPriceCacler struct {
		LenMatch func() bool
		Fee      func() *big.Int
	}
	calcers := []*LenPriceCacler{
		{
			LenMatch: func() bool {
				return len == 5
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(160000)
			},
		},
		{
			LenMatch: func() bool {
				return len == 6
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(80000)
			},
		},
		{
			LenMatch: func() bool {
				return len == 7
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(40000)
			},
		},
		{
			LenMatch: func() bool {
				return len == 8
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(20000)
			},
		},
		{
			LenMatch: func() bool {
				return len == 9
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(10000)
			},
		},
		{
			LenMatch: func() bool {
				return len == 10
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(5000)
			},
		},
		{
			LenMatch: func() bool {
				return len == 11
			},
			Fee: func() *big.Int {
				return params.CoinFromNumer(2500)

			},
		},
		{
			LenMatch: func() bool {
				return len > 11
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

	return drepFee
}

func CheckAlias(tx *types.Transaction, store store.StoreInterface, height uint64) error {
	alias := tx.GetData()
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

	fee := getFee(len(alias))
	from, _ := tx.From()
	originBalance := store.GetBalance(from, height)
	leftBalance := originBalance.Sub(originBalance, fee)
	if leftBalance.Sign() < 0 {
		return fmt.Errorf("set alias ,not enough balance")
	}

	return nil
}
