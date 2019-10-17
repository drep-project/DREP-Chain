package types

import (
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
)

type StakeStorage struct {
	ReceivedCreditAddr  []crypto.CommonAddress //Trust given by oneself and others
	ReceivedCreditValue []big.Int              //value of vote

	//撤销给与别人的信任数据存放于此；等待一段时间或者高度后，value对应的balance加入到Balance中。key是撤销时交易所在的高度
	CancelCreditHeight []uint64
	CancelCreditValue  []big.Int

	CandidateData []byte //注册候选节点时，需要携带的pubkey/ip等信息
}

////存储所有候选人的地址及相关属性信息
//type CandidateInfoStorage struct {
//	Addrs []crypto.CommonAddress
//}
