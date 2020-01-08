package types

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type P2pNode struct {
}

//候选节点数据部分信息
type CandidateData struct {
	Pubkey *secp256k1.PublicKey //出块节点的pubkey
	Node   string               //出块节点的地址
}

func (cd CandidateData) check() error {
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
	items := strings.Split(addr, ":")
	if items == nil || len(items) != 2 {
		return false
	}

	a := net.ParseIP(items[0])
	if a == nil {
		return false
	}

	match, err := regexp.MatchString("^[0-9]*$", items[1])
	if err != nil {
		return false
	}

	i, err := strconv.Atoi(items[1])
	if err != nil {
		return false
	}
	if i < 0 || i > 65535 {
		return false
	}

	if match == false {
		return false
	}

	return true
}
