package database

import (
	"github.com/golang/protobuf/proto"
	"BlockChainTest/bean"
	"errors"
)

var (
	ErrInvalidBytes = errors.New("invalid bytes to unmarshal")
	ErrNoUnmarshallerFound = errors.New("no unmarshaller found")
)

type DBElem interface {
	DBKey() string
	DBMarshal() ([]byte, error)
}

type unmarshaller func([]byte) (DBElem, error)

func blockUnmarshaller(b []byte) (DBElem, error) {
	block := &bean.Block{}
	err := proto.Unmarshal(b, block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

var unmarshallerMap map[byte] unmarshaller

func init() {
	unmarshallerMap = make(map[byte] unmarshaller, 1)
	unmarshallerMap[byte(bean.MsgTypeBlock)] = blockUnmarshaller
}

func marshal(elem DBElem) ([]byte, error) {
	return elem.DBMarshal()
}

func unmarshal(b []byte) (DBElem, error) {
	if b == nil || len(b) < 1 {
		return nil, ErrInvalidBytes
	}
	unmarshaller, ok := unmarshallerMap[b[0]]
	if !ok {
		return nil, ErrNoUnmarshallerFound
	}
	return unmarshaller(b[1:])
}