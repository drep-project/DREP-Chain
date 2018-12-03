package accounts

import (
    "reflect"
    "math/big"
    "errors"
    "encoding/hex"
    "BlockChainTest/mycrypto"
    "BlockChainTest/core/ethhexutil"
    "encoding/json"
)

const (
    HashLength    = 32
    AddressLength = 20
    RootChainID   = 0
)

var (
    ErrExceedHashLength = errors.New("bytes length exceed maximum hash length of 32")
    hashT    = reflect.TypeOf(Hash{})
	addressT = reflect.TypeOf(CommonAddress{})
)

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
        return CommonAddress{}
    }
    return Bytes2Address(b)
}

func (addr CommonAddress) Hex() string {
    return hex.EncodeToString(addr.Bytes())
}

func Big2Address(x *big.Int) CommonAddress {
    return Bytes2Address(x.Bytes())
}

func (addr CommonAddress) Big() *big.Int {
    return new(big.Int).SetBytes(addr.Bytes())
}


// MarshalText returns the hex representation of a.
func (addr CommonAddress) MarshalText() ([]byte, error) {
	return ethhexutil.Bytes(addr[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (addr CommonAddress) UnmarshalText(input []byte) error {
	return ethhexutil.UnmarshalFixedText("Address", input, addr[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (addr *CommonAddress) UnmarshalJSON(input []byte) error {
	return ethhexutil.UnmarshalFixedJSON(addressT, input, addr[:])
}


func PubKey2Address(pubKey *mycrypto.Point) CommonAddress {
    return Bytes2Address(mycrypto.Hash256(pubKey.Bytes()))
}

type ByteCode []byte

func GetByteCodeHash(byteCode ByteCode) Hash {
    return Bytes2Hash(mycrypto.Hash256(byteCode))
}

func GetByteCodeAddress(callerAddr CommonAddress, nonce int64) CommonAddress {
    b, err := json.Marshal(
        struct {
            CallerAddr CommonAddress
            Nonce      int64
        }{
            CallerAddr: callerAddr,
            Nonce:      nonce,
        })
    if err != nil {
        return CommonAddress{}
    }
    return Bytes2Address(mycrypto.Hash256(b))
}

type Hash [HashLength]byte

func Bytes2Hash(b []byte) Hash {
    if b == nil {
        return Hash{}
    }
    var h Hash
    h.SetBytes(b)
    return h
}

func (h Hash) SetBytes(b []byte) {
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