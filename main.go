package main

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    "github.com/drep-project/binary"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "log"
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
    str := "60606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806326121ff01460465780636bf6b038146052575b6000565b3460005760506072565b005b34600057605c6087565b6040518082815260200191505060405180910390f35b6000600081548092919060010191905055505b565b600060005490505b905600a165627a7a72305820087de3f13b992abe1f1e60e25fc4e439883438166ccf424b9ee12509e35e299a0029"
    code, _  := hex.DecodeString(str)
    fmt.Println(code)
    bigValue := &common.Big{}
    valStr := "\"0x9184e72a\""
    err := json.Unmarshal([]byte(valStr), bigValue)
    fmt.Println(err)
    binary.ImportCodeC(reflect.TypeOf(secp256k1.PublicKey{}), &secpPubKeyCodeC{})
    pri, _ := secp256k1.GeneratePrivateKey(nil)
    a := AAA{
        PK:   *pri.PubKey(),
        Name: "xie",
        Age:  123,
        Num:  *big.NewInt(1232131231231231231),
    }
    bytes ,err := binary.Marshal(a)
    if err != nil {
        log.Fatal(err)
    }
    a2 := &AAA{}
    err = binary.Unmarshal(bytes, a2)
    if err != nil {
        log.Fatal(err)
    }
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