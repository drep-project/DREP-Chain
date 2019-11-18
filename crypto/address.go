package crypto

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/common/hexutil"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"math/big"
	"reflect"
)

const (
	HashLength    = 32
	AddressLength = 20
)

var (
	ErrExceedHashLength = errors.New("bytes length exceed maximum hash length of 32")
	hashT               = reflect.TypeOf(Hash{})
	addressT            = reflect.TypeOf(CommonAddress{})
	ZeroAddress         = CommonAddress{}
)

// Address represents the 20 byte address of an Ethereum account.
type CommonAddress [AddressLength]byte

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) CommonAddress {
	var a CommonAddress
	a.SetBytes(b)
	return a
}

// BigToAddress returns Address with byte values of b.
// If b is larger than len(h), b will be cropped from the left.
func BigToAddress(b *big.Int) CommonAddress { return BytesToAddress(b.Bytes()) }

// HexToAddress returns Address with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
func HexToAddress(s string) CommonAddress { return BytesToAddress(common.FromHex(s)) }

func PubkeyToAddress(p *secp256k1.PublicKey) CommonAddress {
	return BytesToAddress(sha3.Keccak256(p.SerializeUncompressed()[1:])[12:])
}

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// Ethereum address or not.
func IsHexAddress(s string) bool {
	if common.HasHexPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*AddressLength && common.IsHex(s)
}

// Bytes gets the string representation of the underlying address.
func (a CommonAddress) Bytes() []byte { return a[:] }

// Hash converts an address to a hash by left-padding it with zeros.
func (a CommonAddress) Hash() Hash { return BytesToHash(a[:]) }

// Hex returns an EIP55-compliant hex string representation of the address.
func (a CommonAddress) Hex() string {
	unchecksummed := hex.EncodeToString(a[:])
	hash := sha3.Keccak256([]byte(unchecksummed))
	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}

// String implements fmt.Stringer.
func (a CommonAddress) String() string {
	return a.Hex()
}

// String implements fmt.Stringer.
func (a CommonAddress) Big() *big.Int {
	return new(big.Int).SetBytes(a[:])
}

// String implements fmt.Stringer.
func (a CommonAddress) IsEmpty() bool {
	return a == ZeroAddress
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (a CommonAddress) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), a[:])
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *CommonAddress) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// MarshalText returns the hex representation of a.
func (a CommonAddress) MarshalText() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (a *CommonAddress) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Address", input, a[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *CommonAddress) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(addressT, input, a[:])
}

type ByteCode []byte

func GetByteCodeHash(byteCode ByteCode) Hash {
	return Bytes2Hash(sha3.Keccak256(byteCode))
}
