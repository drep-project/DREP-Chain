package types

import (
	"bytes"
	"fmt"
	"github.com/drep-project/binary"
	"log"
	"testing"
)

func TestBkMarshal(t *testing.T) {
	var tx = BlockHeader{}
	tx.Version = 1
	bytes1, err := binary.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tx.Hash())
	tx.TxRoot = []byte{1, 2, 3}
	tx.StateRoot = []byte{}
	bytes12, err := binary.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(bytes1, bytes12) {
		log.Fatal("not match marshal result")
	}
}
