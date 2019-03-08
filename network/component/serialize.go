package component

import (
	"errors"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
    "github.com/drep-project/drep-chain/network/types"
	"github.com/vishalkuo/bimap"
	"reflect"
    "encoding/json"
)

var (
	msgTypeMap = bimap.NewBiMap()
)

func Serialize(message interface{}, prvKey *secp256k1.PrivateKey) (*types.Message, error) {
	body, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	refType := reflect.TypeOf(message)
	if refType.Kind() == reflect.Ptr {
		refType = refType.Elem()
	}
	msgType, ok := msgTypeMap.GetInverse(refType)
	if !ok {
		msgTypeMap.GetInverse(refType)
		return nil, errors.New("Unknown peer message type")
	}
	sig, err :=  prvKey.Sign(sha3.Hash256(body))
	if err != nil {
		return nil, err
	}
	msg := &types.Message{
		Header: &types.MessageHeader{
			Type: msgType.(int),
			PubKey: (*secp256k1.PublicKey)(&prvKey.PublicKey),
			Sig: sig,
		},
		Body:body,
	}
	return msg, nil
}

func Deserialize(msgBytes []byte) (interface{}, int, *secp256k1.PublicKey, error) {
	msg := &types.Message{}
	if err := json.Unmarshal(msgBytes, msg); err != nil {
		return nil, 0, nil, err
	}

	refType, ok := msgTypeMap.Get(msg.Header.Type)
	if !ok {
		return nil, 0, nil, errors.New("Unknown peer message type ")
	}
	bodyMsg := reflect.New(refType.(reflect.Type)).Interface()
	if err := json.Unmarshal(msg.Body, bodyMsg); err == nil {
		if !msg.Header.Sig.Verify(sha3.Hash256(msg.Body), msg.Header.PubKey) {
			return nil, 0, nil, errors.New("check signature fail")
		}
		return bodyMsg, msg.Header.Type, msg.Header.PubKey, nil
	} else {
		return nil, 0, nil, err
	}
}

func RegisterMap(msgType int, typeInstance  interface{}) error{
	if msgTypeMap.Exists(msgType) {
		return errors.New("exist type")
	}
	if msgTypeMap.ExistsInverse(msgType) {
		return errors.New("exist instance type")
	}
	msgTypeMap.Insert(msgType, reflect.TypeOf(typeInstance))
	return nil
}