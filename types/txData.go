package types

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/drep-chain/common/hexutil"
	"github.com/drep-project/drep-chain/crypto"
	"net"
	"regexp"
	"strconv"
	"strings"
)

//候选节点数据部分信息
type CandidateData struct {
	P2PPubkey string //出块节点的pubkey
	Addr      string //出块节点的地址
}

func (cd CandidateData) check() error {
	pk,_ := hexutil.Decode(cd.P2PPubkey)

	_, err := crypto.DecompressPubkey(pk)
	if err != nil {
		return fmt.Errorf("pubkey:%s err:%s", cd.P2PPubkey, err.Error())
	}

	if !hostAddrCheck(cd.Addr) {
		return fmt.Errorf("addr err:%s", cd.Addr)
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
