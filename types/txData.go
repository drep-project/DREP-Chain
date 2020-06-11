package types

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/network/p2p/enode"
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
