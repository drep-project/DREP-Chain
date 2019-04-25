package types

import (
	"github.com/drep-project/binary"
	"testing"
)

func TestMarshal(t *testing.T) {
	var tx = TransactionData{}
	tx.Data =  []byte{1,2,3,4,5,6}
	bytes,_ := binary.Marshal(tx)

	block2 := &TransactionData{}
	binary.Unmarshal(bytes, block2)
}