package main

import (
    "fmt"
    "github.com/drep-project/binary"
    "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto/secp256k1"

    "math/big"
    "reflect"
)

type AAA struct {
    PK   secp256k1.PublicKey
    Name string
    Age  int32
    Num  big.Int
}

func main(){
 xxxx := "0x0201002049000a8f1f0fba4503d44f0c207bb9ea0955e700000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080de0b6b3a7640000031e848001c8a8cfa2cb0b00f93b3acab6f8ce3a62f3d19e3416c6c02fd5a06c"
 bytesss := common.MustDecode(xxxx)
    tx := &types.Transaction{}
    err := binary.Unmarshal(bytesss,tx)
    fmt.Println(err)
}

type secpPubKeyCodeC struct{}

// Encode encodes a value into the encoder.
func (c *secpPubKeyCodeC) EncodeTo(e *binary.Encoder, rv reflect.Value) error {
    fff :=  rv.Interface().(secp256k1.PublicKey)
    contents := fff.SerializeCompressed()
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
    rv.Set(reflect.ValueOf(*pk))
    return nil
}