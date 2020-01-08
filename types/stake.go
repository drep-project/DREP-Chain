package types

import (
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/crypto"
	"math/big"
)

//某个高度对应的某creadit value
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
	//ReceivedCreditAddr []crypto.CommonAddress //Trust given by oneself and others
	////ReceivedCreditValue  []big.Int              //value of vote
	////ReceivedCreditHeight []uint64               // height of vote tx
	//ReceivedCreditHeightValue [][]HeightValue

	RC []ReceivedCredit

	//撤销给与别人的信任数据存放于此；
	//CancelCreditHeight []uint64
	//CancelCreditValue  []big.Int
	CC []CancelCredit

	CandidateData []byte //注册候选节点时，需要携带的pubkey/ip等信息
}

type IntersetDetail struct {
	PrincipalData []HeightValue
	IntersetData  []HeightValue
}

func (IntersetDetail) Error() string {
	panic("implement me")
}

//func NewInterestData() *IntersetDetail {
//	return &IntersetDetail{}
//}
//
//func (i *IntersetData)Push(key, value HeightValue)  {
//	i[key] = value
//}
