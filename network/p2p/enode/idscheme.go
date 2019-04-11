// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package enode

import (
	"fmt"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/network/p2p/enr"
	"golang.org/x/crypto/sha3"
	"reflect"
)

// List of known secure identity schemes.
var ValidSchemes = enr.SchemeMap{
	"v4": V4ID{},
}

var ValidSchemesForTesting = enr.SchemeMap{
	"v4":   V4ID{},
	"null": NullID{},
}

// v4ID is the "v4" identity scheme.
type V4ID struct{}

// SignV4 signs a record using the v4 scheme.
func SignV4(r *enr.Record, privkey *secp256k1.PrivateKey) error {
	// Copy r to avoid modifying it if signing fails.
	cpy := *r
	cpy.Set(enr.ID("v4"))

	//fmt.Println(privkey.PublicKey)
	cpy.Set(Secp256k1(privkey.PublicKey))

	h := sha3.NewLegacyKeccak256()
	//rlp.Encode(h, cpy.AppendElements(nil))
	sig, err := crypto.Sign(h.Sum(nil), privkey)
	if err != nil {
		return err
	}
	sig = sig[:len(sig)-1] // remove v
	if err = cpy.SetSig(V4ID{}, sig); err == nil {
		*r = cpy
	}
	return err
}

func (V4ID) Verify(r *enr.Record, sig []byte) error {
	var entry s256raw
	if err := r.Load(&entry); err != nil {
		return err
	} else if len(entry) != 33 {
		return fmt.Errorf("invalid public key")
	}

	h := sha3.NewLegacyKeccak256()
	//rlp.Encode(h, r.AppendElements(nil))
	if !crypto.VerifySignature(entry, h.Sum(nil), sig) {
		return enr.ErrInvalidSig
	}
	return nil
}

func (V4ID) NodeAddr(r *enr.Record) []byte {
	var pubkey Secp256k1
	err := r.Load(&pubkey)
	if err != nil {
		return nil
	}
	//buf := make([]byte, 32)
	//math.ReadBits(pubkey.X, buf[:32])
	//math.ReadBits(pubkey.Y, buf[32:])
	//return crypto.Keccak256(buf)
	return  crypto.CompressPubkey(pubkey.secp2561PublicKey())[1:]
}


// Secp256k1 is the "secp256k1" key, which holds a public key.
type Secp256k1 secp256k1.PublicKey

func (v Secp256k1) secp2561PublicKey() *secp256k1.PublicKey{
	pk := (secp256k1.PublicKey)(v)
	return &pk
}


func (v Secp256k1) ENRKey() string { return "secp256k1" }

func init(){
	binary.ImportCodeC(reflect.TypeOf(Secp256k1{}), &secpPubKeyCodeC{})
}

type secpPubKeyCodeC struct{}

// Encode encodes a value into the encoder.
func (c *secpPubKeyCodeC) EncodeTo(e *binary.Encoder, rv reflect.Value) error {
	pk :=  secp256k1.PublicKey(rv.Interface().(Secp256k1))
	contents := pk.SerializeCompressed()
	e.WriteUvarint(uint64(len(contents)))
	e.Write(contents)
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *secpPubKeyCodeC) DecodeTo(d *binary.Decoder, rv reflect.Value) (err error) {
	length, err := d.ReadUvarint()
	if err != nil {
		return err
	}
	contents := make([]byte, length)
	_, err = d.Read(contents)
	if err != nil {
		return err
	}
	pk, err := secp256k1.ParsePubKey(contents)
	if err != nil {
		return err
	}
	val := Secp256k1(*pk)
	rv.Set(reflect.ValueOf(val))
	return nil
}


// EncodeRLP implements rlp.Encoder.
//func (v Secp256k1) EncodeRLP(w io.Writer) error {
//	return rlp.Encode(w, crypto.CompressPubkey((*secp256k1.PublicKey)(&v)))
//}

// DecodeRLP implements rlp.Decoder.
//func (v *Secp256k1) DecodeRLP(s *rlp.Stream) error {
//	buf, err := s.Bytes()
//	if err != nil {
//		return err
//	}
//	pk, err := crypto.DecompressPubkey(buf)
//	if err != nil {
//		return err
//	}
//	*v = (Secp256k1)(*pk)
//	return nil
//}

// s256raw is an unparsed secp256k1 public key entry.
type s256raw []byte

func (s256raw) ENRKey() string { return "secp256k1" }

// v4CompatID is a weaker and insecure version of the "v4" scheme which only checks for the
// presence of a secp256k1 public key, but doesn't verify the signature.
type v4CompatID struct {
	V4ID
}

func (v4CompatID) Verify(r *enr.Record, sig []byte) error {
	var pubkey Secp256k1
	return r.Load(&pubkey)
}

func signV4Compat(r *enr.Record, pubkey *secp256k1.PublicKey) {
	r.Set((*Secp256k1)(pubkey))
	if err := r.SetSig(v4CompatID{}, []byte{}); err != nil {
		panic(err)
	}
}

// NullID is the "null" ENR identity scheme. This scheme stores the node
// ID in the record without any signature.
type NullID struct{}

func (NullID) Verify(r *enr.Record, sig []byte) error {
	return nil
}

func (NullID) NodeAddr(r *enr.Record) []byte {
	var id ID
	r.Load(enr.WithEntry("nulladdr", &id))
	return id[:]
}

func SignNull(r *enr.Record, id ID) *Node {
	r.Set(enr.ID("null"))
	r.Set(enr.WithEntry("nulladdr", id))
	if err := r.SetSig(NullID{}, []byte{}); err != nil {
		panic(err)
	}
	return &Node{r: *r, id: id}
}
