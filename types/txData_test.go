package types

import (
	"fmt"
	"testing"
)

func TestCandidateMarshal(t *testing.T)  {
	//p, _ := crypto.GenerateKey(rand.Reader)
	//b,_:= p.PubKey().MarshalText()
	//fmt.Println(string(b))
	//
	//
	////}
	////_, err := secp256k1.ParsePubKey([]byte("0x0373654ccdb250f2cfcfe64c783a44b9ea85bc47f2f00c480d05082428d277d6d0"))
	//pk := secp256k1.NewPublicKey(nil,nil)
	//err := pk.UnmarshalText([]byte(cd.Pubkey))


	cd := &CandidateData{
		Pubkey: string("0x0373654ccdb250f2cfcfe64c783a44b9ea85bc47f2f00c480d05082428d277d6d0"),
		Addr:   "127.0.0.12:34",
	}

	b,err := cd.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	//fmt.Println(string(b))

	cd2 := &CandidateData{}
	err = cd2.Unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(cd2)

	//50f36097546be34dceae65e65b36f300012c348d2c43e751c33533007be1c9f5
}
