package types

import (
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
)

type StakeStorage struct {
	ReceivedVoteCredit map[crypto.CommonAddress]big.Int //Trust given by oneself and others
	//SentVoteCredit     map[crypto.CommonAddress]big.Int //Vote for Trust to Address
	CancelVoteCredit map[uint64]big.Int //撤销给与别人的信任数据存放于此；等待一段时间或者高度后，value对应的balance加入到Balance中。key是撤销时交易所在的高度

	//LockBalance        map[crypto.CommonAddress]big.Int //智能合约中即可
}
