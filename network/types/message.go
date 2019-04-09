package types

import (
    "github.com/drep-project/drep-chain/crypto/secp256k1"
)


type MessageHeader struct {
    Type int
    //Size int32
    PubKey *secp256k1.PublicKey
    Sig *secp256k1.Signature
}

type Message struct {
    Header *MessageHeader
    Body   []byte
}