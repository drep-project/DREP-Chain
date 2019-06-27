package common

import (
	"fmt"

	"github.com/drep-project/binary"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"math/big"
	"testing"
)

type A struct {
	AA *Big
}

func TestBsonBig(t *testing.T) {
	bigInt := *big.NewInt(123123)
	bigv := Big(bigInt)
	a1 := A{&bigv}
	tx, err := bson.Marshal(a1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(tx))
	a2 := A{}
	err = bson.Unmarshal(tx, &a2)
	if err != nil {
		log.Fatal(err)
	}
	if bigInt.Uint64() != a2.AA.ToInt().Uint64() {
		t.Error("commom big not matches")
	}
}

func TestBinaryBig(t *testing.T) {
	bigInt := *big.NewInt(123123)
	bigv := Big(bigInt)
	a1 := A{&bigv}
	tx, err := binary.Marshal(a1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(tx))
	a2 := A{}
	err = binary.Unmarshal(tx, &a2)
	if err != nil {
		log.Fatal(err)
	}
	if bigInt.Uint64() != a2.AA.ToInt().Uint64() {
		t.Error("commom big not matches")
	}
}
