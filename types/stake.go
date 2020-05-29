package types

import (
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/crypto"
	"math/big"
)

//A creadit value of some height
type HeightValue struct {
	CreditHeight uint64
	CreditValue  common.Big
}

type ReceivedCredit struct {
	Addr        crypto.CommonAddress
	HeghtValues []HeightValue
}

type CancelCredit struct {
	CancelCreditHeight uint64
	CancelCreditValue  []big.Int
}

type StakeStorage struct {
	RC []ReceivedCredit

	//The data of revoking the trust given to others is stored here;
	CC []CancelCredit

	// when registering candidate nodes, you need to carry pubkey/ IP and other information
	CandidateData []byte
}

type CancelCreditDetail struct {
	PrincipalData []HeightValue
}

func (CancelCreditDetail) Error() string {
	panic("implement me")
}

//func NewInterestData() *CancelCreditDetail {
//	return &CancelCreditDetail{}
//}
//
//func (i *IntersetData)Push(key, value HeightValue)  {
//	i[key] = value
//}
