package crypto

import (
	"github.com/drep-project/drep-chain/common"
	"math/big"
)

type Hash [HashLength]byte

func Bytes2Hash(b []byte) Hash {
	if b == nil {
		return Hash{}
	}
	var h Hash
	h.SetBytes(b)
	return h
}

func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		panic(ErrExceedHashLength)
	}
	copy(h[HashLength-len(b):], b)
}

func (h Hash) Bytes() []byte {
	b := make([]byte, len(h))
	copy(b, h[:])
	return b
}
func (h Hash) String() string {
	strBytes, _ := common.Bytes(h[:]).MarshalText()
	return string(strBytes)
}

func Big2Hash(x *big.Int) Hash {
	if x == nil {
		return Hash{}
	}
	return Bytes2Hash(x.Bytes())
}

// UnmarshalText parses a hash in hex syntax.
func (h *Hash) UnmarshalText(input []byte) error {
	return common.UnmarshalFixedText("Hash", input, h[:])
}

// IsEqual returns true if target is the same as hash.
func (hash *Hash) IsEqual(target *Hash) bool {
	if hash == nil && target == nil {
		return true
	}
	if hash == nil || target == nil {
		return false
	}
	return *hash == *target
}

// UnmarshalJSON parses a hash in hex syntax.
func (h *Hash) UnmarshalJSON(input []byte) error {
	return common.UnmarshalFixedJSON(hashT, input, h[:])
}

// MarshalText returns the hex representation of h.
func (h Hash) MarshalText() ([]byte, error) {
	return common.Bytes(h[:]).MarshalText()
}
