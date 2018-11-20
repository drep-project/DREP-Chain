package bean

import (
	"encoding/hex"
	"math/big"
	"errors"
	"encoding/json"
	"BlockChainTest/mycrypto"
)

const (
	HashLength = 32
	AddressLength = 20
)

var (
	ErrExceedHashLength = errors.New("bytes length exceed maximum hash length of 32")
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

func Big2Hash(x *big.Int) Hash {
	if x == nil {
		return Hash{}
	}
	return Bytes2Hash(x.Bytes())
}

type CommonAddress [AddressLength]byte

func (addr CommonAddress) IsEmpty() bool {
	return addr == CommonAddress{}
}

func Bytes2Address(b []byte) CommonAddress {
	if b == nil {
		return CommonAddress{}
	}
	var addr CommonAddress
	addr.SetBytes(b)
	return addr
}

func (addr *CommonAddress) SetBytes(b []byte) {
	if len(b) > len(addr) {
		copy(addr[:], b[len(b) - AddressLength:])
	} else {
		copy(addr[AddressLength-len(b):], b)
	}
}

func (addr CommonAddress) Bytes() []byte {
	return addr[:]
}

func Hex2Address(s string) CommonAddress {
	if s == "" {
		return CommonAddress{}
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return Bytes2Address(b)
}

func (addr CommonAddress) Hex() string {
	return hex.EncodeToString(addr.Bytes())
}

func (addr CommonAddress) Big() *big.Int {
	return new(big.Int).SetBytes(addr.Bytes())
}

func PubKey2Address(pubKey *mycrypto.Point) CommonAddress {
	return Bytes2Address(mycrypto.Hash256(pubKey.Bytes()))
}

type Account struct {
	Addr                 CommonAddress
	Nonce                int64
	Balance              *big.Int
	IsContract           bool
	ByteCode             []byte
	CodeHash             []byte
}

func MarshalAccount(account *Account) ([]byte, error) {
	b, err := json.Marshal(account)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func UnmarshalAccount(b []byte) (*Account, error) {
	account := &Account{}
	err := json.Unmarshal(b, account)
	if err != nil {
		return nil, err
	}
	return account, nil
}