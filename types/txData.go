package types

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/network/p2p/enode"
)

//候选节点数据部分信息
type CandidateData struct {
	Pubkey *secp256k1.PublicKey //出块节点的pubkey
	Node   string               //出块节点的地址
}

func (cd CandidateData) check() error {
	if !checkp2pNode(cd.Node) {
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

func checkp2pNode(node string) bool {
	n := enode.Node{}
	return n.UnmarshalText([]byte(node)) == nil
}
