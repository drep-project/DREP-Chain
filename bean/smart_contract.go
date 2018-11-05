package bean

import (
	"encoding/hex"
	"math/big"
	"errors"
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
		copy(addr[:], b[len(b) - AddressLen:])
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

func Big2Address(x *big.Int) CommonAddress {
	if x == nil {
		return CommonAddress{}
	}
	return Bytes2Address(x.Bytes())
}

func (addr CommonAddress) Big() *big.Int {
	return new(big.Int).SetBytes(addr.Bytes())
}

func (addr CommonAddress) IsEmpty() bool {
	return addr == CommonAddress{}
}

//type ByteCode []byte
//
//func CodeHash(byteCode ByteCode) Hash {
//	b := mycrypto.Hash256(byteCode)
//	return Bytes2Hash(b)
//}
//
//func CodeAddr(byteCode ByteCode) CommonAddress {
//	return Bytes2Address(mycrypto.Hash256(byteCode))
//}
//
//func Bytes2Key(b []byte) string {
//	return hex.EncodeToString(mycrypto.Hash256(b))
//}
//
//func Address2Key(addr CommonAddress) string {
//	return Bytes2Key(addr.Bytes())
//}
//
//func (account *ContractAccount) DBKey() string {
//	return Bytes2Key(account.Addr)
//}
//
//func (account *ContractAccount) DBMarshal() ([]byte, error) {
//	_b, err := proto.Marshal(account)
//	if err != nil {
//		return nil, err
//	}
//	b := make([]byte, len(_b) + 1)
//	b[0] = byte(MsgTypeContractAccount)
//	copy(b[1:], _b)
//	return b, nil
//}
//
//func (account *ContractAccount) Address() CommonAddress {
//	return Bytes2Address(account.Addr)
//}
//
//func (account *ContractAccount) Exists() bool {
//	return !account.Address().IsEmpty()
//}
//
//func (log *Log) DBKey() string {
//	return Address2Key(log.Address())
//}
//
//func (log *Log) DBMarshal() ([]byte, error) {
//	_b, err := proto.Marshal(log)
//	if err != nil {
//		return nil, err
//	}
//	b := make([]byte, len(_b) + 1)
//	b[0] = byte(MsgTypeLog)
//	copy(b[1:], _b)
//	return b, nil
//}
//
//func (log *Log) Address() CommonAddress {
//	b := make([]byte, len(log.ContractAddr) + len(log.TxHash))
//	copy(b, log.ContractAddr)
//	copy(b[len(log.ContractAddr):], log.TxHash)
//	b = mycrypto.Hash256(b)
//	return Bytes2Address(b)
//}
//
//func (log *Log) Exists() bool {
//	return !log.Address().IsEmpty()
//}