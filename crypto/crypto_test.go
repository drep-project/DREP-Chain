// Copyright 2018 DREP Foundation Ltd.
// This file is part of the drep-cli library.
//
// The drep-cli library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The drep-cli library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the drep-cli library. If not, see <http://www.gnu.org/licenses/>.

package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/drep-project/DREP-Chain/common/hexutil"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"math/big"
	"testing"

	"github.com/drep-project/DREP-Chain/common"
)

var testAddrHex = "0x0b105f01e19fa1683f40ff86f8c9d4381fdc9631"
var testPrivHex = "0x289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"

// These tests are sanity checks.
// They should ensure that we don't e.g. use Sha3-224 instead of Sha3-256
// and that the sha3 library uses keccak-f permutation.
func TestKeccak256Hash(t *testing.T) {
	msg := []byte("abc")
	exp, _ := hex.DecodeString("4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45")
	checkhash(t, "Sha3-256-array", func(in []byte) []byte { h := Keccak256Hash(in); return h[:] }, msg, exp)
}

func TestToECDSAErrors(t *testing.T) {
	if _, err := FromPrivString("0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Fatal("ethcrypto.FromPrivString should've returned error")
	}
	if _, err := FromPrivString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"); err == nil {
		t.Fatal("ethcrypto.FromPrivString should've returned error")
	}
}

func BenchmarkSha3(b *testing.B) {
	a := []byte("hello world")
	for i := 0; i < b.N; i++ {
		sha3.Keccak256(a)
	}
}

func TestSign(t *testing.T) {

	key, _ := FromPrivString(testPrivHex)

	addr := HexToAddress(testAddrHex)

	msg := sha3.Keccak256([]byte("foo"))
	sig, err := Sign(msg, key)
	if err != nil {
		t.Errorf("Sign error: %s", err)
	}
	recoveredPub, err := Ecrecover(msg, sig, false)
	if err != nil {
		t.Errorf("ECRecover error: %s", err)
	}

	pubKey, _ := secp256k1.ParsePubKey(recoveredPub)
	recoveredAddr := PubkeyToAddress(pubKey)
	if addr != recoveredAddr {
		t.Errorf("GetAddress mismatch: want: %x have: %x", addr, recoveredAddr)
	}

	// should be equal to SigToPub
	recoveredPub2, err := SigToPub(msg, sig)
	if err != nil {
		t.Errorf("ECRecover error: %s", err)
	}
	recoveredAddr2 := PubkeyToAddress(recoveredPub2)
	if addr != recoveredAddr2 {
		t.Errorf("GetAddress mismatch: want: %x have: %x", addr, recoveredAddr2)
	}
}

func TestInvalidSign(t *testing.T) {
	if _, err := Sign(make([]byte, 1), nil); err == nil {
		t.Errorf("expected sign with hash 1 byte to error")
	}
	if _, err := Sign(make([]byte, 33), nil); err == nil {
		t.Errorf("expected sign with hash 33 byte to error")
	}
}

func TestNewContractAddress(t *testing.T) {
	key, _ := FromPrivString(testPrivHex)
	addr := HexToAddress(testAddrHex)
	genAddr := PubkeyToAddress(key.PubKey())
	// sanity check before using addr to create contract address
	checkAddr(t, genAddr, addr)

	caddr0 := CreateAddress(addr, 0)
	caddr1 := CreateAddress(addr, 1)
	caddr2 := CreateAddress(addr, 2)

	checkAddr(t, HexToAddress("38bd4b79a96b5ddc536639cc0cc927c7ee5511bc"), caddr0)
	checkAddr(t, HexToAddress("7e11723d6f9a0601f0343f96cdbd01e303b8ee9a"), caddr1)
	checkAddr(t, HexToAddress("39b35935d5417dad741f45d9f39e20596bada9d3"), caddr2)
}

func TestValidateSignatureValues(t *testing.T) {
	check := func(expected bool, v byte, r, s *big.Int) {
		if ValidateSignatureValues(v, r, s, false) != expected {
			t.Errorf("mismatch for v: %d r: %d s: %d want: %v", v, r, s, expected)
		}
	}
	minusOne := big.NewInt(-1)
	one := common.Big1
	zero := common.Big0
	secp256k1nMinus1 := new(big.Int).Sub(secp256k1N, common.Big1)

	// correct v,r,s
	check(true, 0, one, one)
	check(true, 1, one, one)
	// incorrect v, correct r,s,
	check(false, 2, one, one)
	check(false, 3, one, one)

	// incorrect v, combinations of incorrect/correct r,s at lower limit
	check(false, 2, zero, zero)
	check(false, 2, zero, one)
	check(false, 2, one, zero)
	check(false, 2, one, one)

	// correct v for any combination of incorrect r,s
	check(false, 0, zero, zero)
	check(false, 0, zero, one)
	check(false, 0, one, zero)

	check(false, 1, zero, zero)
	check(false, 1, zero, one)
	check(false, 1, one, zero)

	// correct sig with max r,s
	check(true, 0, secp256k1nMinus1, secp256k1nMinus1)
	// correct v, combinations of incorrect r,s at upper limit
	check(false, 0, secp256k1N, secp256k1nMinus1)
	check(false, 0, secp256k1nMinus1, secp256k1N)
	check(false, 0, secp256k1N, secp256k1N)

	// current callers ensures r,s cannot be negative, but let's test for that too
	// as crypto package could be used stand-alone
	check(false, 0, minusOne, one)
	check(false, 0, one, minusOne)
}

func checkhash(t *testing.T, name string, f func([]byte) []byte, msg, exp []byte) {
	sum := f(msg)
	if !bytes.Equal(exp, sum) {
		t.Fatalf("hash %s mismatch: want: %x have: %x", name, exp, sum)
	}
}

func checkAddr(t *testing.T, addr0, addr1 CommonAddress) {
	if addr0 != addr1 {
		t.Fatalf("address mismatch: want: %x have: %x", addr0, addr1)
	}
}

func FromPrivString(str string) (*secp256k1.PrivateKey, error) {
	bytes, err := common.Decode(str)
	if err != nil {
		return nil, err
	}
	prikey, _ := ToPrivateKey(bytes)
	return prikey, nil
}

func TestPubkey(t *testing.T) {
	//stmp, _ := GeneratePrivateKey(rand.Reader)
	b, _ := hexutil.Decode("0x03ad000bc9a4a29c11227d6b5ee05076b594c1c22567cdd85fbb8222e6a715924e")
	pk, err := DecompressPubkey(b)
	//crypto.FromECDSAPub(&key)[1:]
	c := PubkeyToAddress(pk)
	bb, _ := c.MarshalText()
	fmt.Println(string(bb))
	c2 := &CommonAddress{}
	err = c2.UnmarshalJSON(bb)
	//fmt.Println(err)
	//fmt.Println(c.Hex())

	a := hexutil.Encode(c[:])
	//a := fmt.Sprintf("%x", c)
	fmt.Println(err, a)
}

func TestPAddrJson(t *testing.T) {
	addr := `"0xe91f67944ec2f7223bf6d0272557a5b13ecc1f28"`
	ca := &CommonAddress{}

	//addrBuf, err := hexutil.Decode(addr)
	//ca.SetBytes(addrBuf)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//buf, err := ca.MarshalText()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println(string(buf))

	//ca1 := &CommonAddress{}
	//err := ca.UnmarshalText([]byte(addr))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println(ca.Hex())

	err := ca.UnmarshalJSON([]byte(addr))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ca.Hex())
}

func TestDrepToEth(t *testing.T) {
	fmt.Println("eth:", DrepToEth("0x3726430e6E3448753F29E1d3ca24120850433Ff9").String())

	addr := DrepToEth("drep3726430e6E3448753F29E1d3ca24120850433Ff9")

	fmt.Println("drep:", EthToDrep(&addr))
}
