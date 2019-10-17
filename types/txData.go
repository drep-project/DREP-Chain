package types

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/drep-chain/common/hexutil"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/network/p2p/enode"
)

type P2pNode struct {

}
//候选节点数据部分信息
type CandidateData struct {
	Pubkey string       //出块节点的pubkey
	Node  string //出块节点的地址
}

func (cd CandidateData) check() error {
	pk,_ := hexutil.Decode(cd.Pubkey)

	_, err := crypto.DecompressPubkey(pk)
	if err != nil {
		return fmt.Errorf("pubkey:%s err:%s", cd.Pubkey, err.Error())
	}

	if !hostAddrCheck(cd.Node) {
		return fmt.Errorf("addr err:%s", cd.Node)
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

func hostAddrCheck(addr string) bool {
	n := enode.Node{}
	err := n.UnmarshalText([]byte(addr))
	return err != nil
}
