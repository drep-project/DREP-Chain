package main

import (
    "errors"
    "fmt"
    "github.com/drep-project/binary"
    "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "time"

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

 xxxx := "0x02cca975000a5a79a051afa6d526653440ab6988db65cd54f5000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000502540be4000504a817c80002ea60e2ebc0cb0b00c4ac59f52b3052e5c14566ed397453ea913c6fbc4120884f8e4d26c77b82e07568ba6cca026063b484883f3fdd8a43d47c4ae44b4f6339480674bc7b2b683a3f5561d3c922a174f127a6678d1eb627ad87fbdf74468e"
 bytesss := common.MustDecode(xxxx)
    tx := &types.Transaction{}
    err := binary.Unmarshal(bytesss,tx)

    fmt.Println(time.Now())
    for i:=0; i<5000;i++{
        verify(tx)
    }
    fmt.Println(time.Now())
    fmt.Println(err)
}

func verify(tx *types.Transaction) (bool, error) {
    if tx.Sig != nil {
        pk, _, err := secp256k1.RecoverCompact(tx.Sig, tx.TxHash().Bytes())
        if err != nil {
            return false, err
        }
        sig := secp256k1.RecoverSig(tx.Sig)
        isValid := sig.Verify(tx.TxHash().Bytes(), pk)
        if err != nil {
            return false, err
        }
        if !isValid {
            return false, errors.New("signature not validate")
        }
        return true, nil
    } else {
        return false, errors.New("must assign a signature for transaction")
    }
 return true, nil
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