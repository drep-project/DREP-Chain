package types

import (
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/francoispqt/gojay"
)


type MessageHeader struct {
    Type int
    //Size int32
    PubKey *secp256k1.PublicKey
    Sig *secp256k1.Signature
}

// UnmarshalJSONObject implements gojay's UnmarshalerJSONObject
func (v *MessageHeader) UnmarshalJSONObject(dec *gojay.Decoder, k string) error {
    switch k {
    case "type":
        return dec.Int(&v.Type)
    case "pubkey":
        if v.PubKey == nil {
            v.PubKey = &secp256k1.PublicKey{}
        }
        return dec.Object(v.PubKey)
    case "sig":
        if v.Sig == nil {
            v.Sig = &secp256k1.Signature{}
        }
        return dec.Object(v.Sig)
    }
    return nil
}

// NKeys returns the number of keys to unmarshal
func (v *MessageHeader) NKeys() int { return 3 }

// MarshalJSONObject implements gojay's MarshalerJSONObject
func (v *MessageHeader) MarshalJSONObject(enc *gojay.Encoder) {
    enc.IntKey("type", v.Type)
    enc.ObjectKey("pubkey", v.PubKey)
    enc.ObjectKey("sig", v.Sig)
}

// IsNil returns wether the structure is nil value or not
func (v *MessageHeader) IsNil() bool { return v == nil }

type Message struct {
    Header *MessageHeader
    Body   []byte
}

// UnmarshalJSONObject implements gojay's UnmarshalerJSONObject
func (v *Message) UnmarshalJSONObject(dec *gojay.Decoder, k string) error {
    switch k {
    case "header":
        if v.Header == nil {
            v.Header = &MessageHeader{}
        }
        return dec.Object(v.Header)
    case "body":
        body := ""
        err := dec.String(&body)
        v.Body = []byte(body)
        return err
    }

    return nil
}

// NKeys returns the number of keys to unmarshal
func (v *Message) NKeys() int { return 2 }

// MarshalJSONObject implements gojay's MarshalerJSONObject
func (v *Message) MarshalJSONObject(enc *gojay.Encoder) {
    enc.ObjectKey("header", v.Header)
    enc.StringKey("body", string(v.Body))
}

// IsNil returns wether the structure is nil value or not
func (v *Message) IsNil() bool { return v == nil }
