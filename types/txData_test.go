package types

import (
	"crypto/rand"
	"fmt"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/crypto"
	"testing"
)

func TestCandidateMarshal(t *testing.T) {
	p, _ := crypto.GenerateKey(rand.Reader)

	cd := &CandidateData{
		Pubkey: p.PubKey(),
		Node:   "enode://e77d64fecbb1c7e78231507fdd58c963cdc1e0ed0bec29b5a65de32b992d596f@149.129.172.91:44444",
	}

	b, err := cd.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))

	cb := common.Bytes(b)

	cbb, _ := cb.MarshalText()
	fmt.Println(string(cbb))

	cd2 := &CandidateData{}
	err = cd2.Unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(cd2)

	//50f36097546be34dceae65e65b36f300012c348d2c43e751c33533007be1c9f5
}
