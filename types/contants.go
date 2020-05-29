package types

import "math/big"

type TxType uint64

const (
	TransferType TxType = iota

	CreateContractType
	CallContractType

	SetAliasType         //Give the address a nickname
	VoteCreditType       //Pledge to someone else
	CancelVoteCreditType //Revocation of pledge currency
	CandidateType        //Apply to be a candidate block node
	CancelCandidateType  //Apply to be a candidate block node
	RegisterProducer
)

var (
	TransferGas         = big.NewInt(30000)
	MinerGas            = big.NewInt(20000)
	CreateContractGas   = big.NewInt(1000000)
	CallContractGas     = big.NewInt(10000000)
	CrossChainGas       = big.NewInt(10000000)
	SeAliasGas          = big.NewInt(10000000)
	RegisterProducerGas = big.NewInt(10000000)
)
