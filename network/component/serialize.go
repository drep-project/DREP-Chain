package component

import (
	"encoding/json"
	"errors"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/vishalkuo/bimap"
	"reflect"
)
const (
	MsgTypeBlockHeader = iota
	MsgTypeBlock
	MsgTypeTransaction
	MsgTypeSetUp
	MsgTypeCommitment
	MsgTypeChallenge
	MsgTypeResponse
	MsgTypeFail
	MsgTypeNewPeer
	MsgTypePeerList
	MsgTypeBlockReq   //10
	MsgTypeBlockResp
	MsgTypePing
	MsgTypePong
	MsgTypeOfflinePeers
	MsgTypeFirstPeerInfoList
	MsgTypePeerState
	MsgTypeReqPeerState
)

var (
	msgTypeMap = bimap.NewBiMap()
)

func init(){
	// register should done in every module that  needs
	/*
	msgTypeMap.Insert(MsgTypeBlockHeader, reflect.TypeOf(bean.BlockHeader{}))
	msgTypeMap.Insert(MsgTypeBlock, reflect.TypeOf(bean.Block{}))
	msgTypeMap.Insert(MsgTypeTransaction, reflect.TypeOf(bean.Transaction{}))
	msgTypeMap.Insert(MsgTypeSetUp, reflect.TypeOf(bean.Setup{}))
	msgTypeMap.Insert(MsgTypeCommitment, reflect.TypeOf(bean.Commitment{}))
	msgTypeMap.Insert(MsgTypeChallenge, reflect.TypeOf(bean.Challenge{}))
	msgTypeMap.Insert(MsgTypeResponse, reflect.TypeOf(bean.Response{}))
	msgTypeMap.Insert(MsgTypeFail, reflect.TypeOf(bean.Fail{}))
	msgTypeMap.Insert(MsgTypeNewPeer, reflect.TypeOf(bean.PeerInfo{}))
	msgTypeMap.Insert(MsgTypePeerList, reflect.TypeOf(bean.PeerInfoList{}))
	msgTypeMap.Insert(MsgTypeBlockReq, reflect.TypeOf(bean.BlockReq{}))
	msgTypeMap.Insert(MsgTypeBlockResp, reflect.TypeOf(bean.BlockResp{}))
	msgTypeMap.Insert(MsgTypePing, reflect.TypeOf(bean.Ping{}))
	msgTypeMap.Insert(MsgTypePong, reflect.TypeOf(bean.Pong{}))
	msgTypeMap.Insert(MsgTypeOfflinePeers, reflect.TypeOf(bean.OfflinePeers{}))
	msgTypeMap.Insert(MsgTypeFirstPeerInfoList, reflect.TypeOf(bean.FirstPeerInfoList{}))
	msgTypeMap.Insert(MsgTypePeerState, reflect.TypeOf(PeerState{}))
	msgTypeMap.Insert(MsgTypeReqPeerState, reflect.TypeOf(ReqPeerState{}))
	*/
}

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

func GenerateMessage(message interface{}, prvKey *secp256k1.PrivateKey) (*Message, error) {
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
		return nil, errors.New("Unknown peer message type ")
	}
	sig, err :=  prvKey.Sign(sha3.Hash256(body))
	if err != nil {
		return nil, err
	}
	msg := &Message{
		Header: &MessageHeader{
			Type: msgType.(int),
			PubKey: (*secp256k1.PublicKey)(&prvKey.PublicKey),
			Sig: sig,
		},
		Body:body,
	}
	return msg, nil
}

func GetMessage(msgBytes []byte) (interface{}, int, *secp256k1.PublicKey, error) {
	msg := &Message{}
	if err := json.Unmarshal(msgBytes, msg); err != nil {
		return nil, 0, nil, err
	}

	refType, ok := msgTypeMap.Get(msg.Header.Type)
	if !ok {
		return nil, 0, nil, errors.New("Unknown peer message type ")
	}
	bodyMsg := reflect.New(refType.(reflect.Type)).Interface()
	if err := json.Unmarshal(msg.Body, bodyMsg); err == nil {
		if !msg.Header.Sig.Verify( sha3.Hash256(msg.Body), msg.Header.PubKey) {
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