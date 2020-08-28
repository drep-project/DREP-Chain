package types

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/network/p2p/enode"
	"github.com/drep-project/DREP-Chain/params"
)

//Candidate node data section information
type CandidateData struct {
	Pubkey *secp256k1.PublicKey //The pubkey of Candidate node
	Node   string               //address of Candidate node
}

func (cd CandidateData) check() error {
	if !checkp2pNode(cd.Node) {
		return fmt.Errorf("node err:%s", cd.Node)
	}

	return nil
}

func (cd *CandidateData) Marshal() ([]byte, error) {
	err := cd.check()
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(cd)
	return b, nil
}

func (cd *CandidateData) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, cd)
	if err != nil {
		return err
	}

	return cd.check()
}

func checkp2pNode(node string) bool {
	n := enode.Node{}
	return n.UnmarshalText([]byte(node)) == nil
}

func GetAliasFee(len int) *big.Int {
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

func CheckAlias(alias []byte) (*big.Int, error) {
	if len(alias) < 5 {
		return new(big.Int), fmt.Errorf("alias too short")
	}
	if len(alias) > 20 {
		return new(big.Int), fmt.Errorf("alias too long")
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
		return new(big.Int), fmt.Errorf("alias only support number and letter")
	}

	fee := GetAliasFee(len(alias))
	//from, _ := tx.From()
	//originBalance := store.GetBalance(from, height)
	//leftBalance := originBalance.Sub(originBalance, fee)
	//if leftBalance.Sign() < 0 {
	//	return new(big.Int), fmt.Errorf("set alias ,not enough balance")
	//}

	return fee, nil
}
