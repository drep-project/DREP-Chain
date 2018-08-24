package bean

import (
	"github.com/golang/protobuf/proto"
	"errors"
)

func Serialize(message interface{}) (*Serializable, error) {
	msg, ok := message.(proto.Message);
	if !ok {
		return nil, errors.New("bad message type")
	}
	body, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	serializable := &Serializable{Body: body}
	switch message.(type) {
	case *Point:
		serializable.Header = MessageHeader_POINT
	case *PrivateKey:
		serializable.Header = MessageHeader_PRIVATE_KEY
	case *Signature:
		serializable.Header = MessageHeader_SIGNATURE
	case *Announcement:
		serializable.Header = MessageHeader_ANNOUNCEMENT
	case *Commitment:
		serializable.Header = MessageHeader_COMMITMENT
	case *Challenge:
		serializable.Header = MessageHeader_CHALLENGE
	case *Response:
		serializable.Header = MessageHeader_RESPONSE
	case *BlockHeader:
		serializable.Header = MessageHeader_BLOCK_HEADER
	case *TransactionData:
		serializable.Header = MessageHeader_TRANSACTION_DATA
	default:
		return nil, errors.New("bad message type")
	}
	return serializable, nil
}

func Deserialize(msg []byte) (*Serializable, interface{}, error) {
	serializable := &Serializable{}
	if err := proto.Unmarshal(msg, serializable); err != nil {
		return nil, nil, err
	}
	body := serializable.GetBody()
	switch serializable.GetHeader() {
	case MessageHeader_POINT:
		point := &Point{}
		if err := proto.Unmarshal(body, point); err == nil {
			return serializable, point, nil
		} else {
			return nil, nil, err
		}
	case MessageHeader_PRIVATE_KEY:
		prvKey := &PrivateKey{}
		if err := proto.Unmarshal(body, prvKey); err == nil {
			return serializable, prvKey, nil
		} else {
			return nil, nil, err
		}
	case MessageHeader_SIGNATURE:
		sig := &Signature{}
		if err := proto.Unmarshal(body, sig); err == nil {
			return serializable, sig, nil
		} else {
			return nil, nil, err
		}
	case MessageHeader_ANNOUNCEMENT:
		announcement := &Announcement{}
		if err := proto.Unmarshal(body, announcement); err == nil {
			return serializable, announcement, nil
		} else {
			return nil, nil, err
		}
	case MessageHeader_COMMITMENT:
		commitment := &Commitment{}
		if err := proto.Unmarshal(body, commitment); err == nil {
			return serializable, commitment, nil
		} else {
			return nil, nil, err
		}
	case MessageHeader_CHALLENGE:
		challenge := &Challenge{}
		if err := proto.Unmarshal(body, challenge); err == nil {
			return serializable, challenge, nil
		} else {
			return nil, nil, err
		}
	case MessageHeader_RESPONSE:
		response := &Response{}
		if err := proto.Unmarshal(body, response); err == nil {
			return serializable, response, nil
		} else {
			return nil, nil, err
		}
	case MessageHeader_BLOCK_HEADER:
		blockHeader := &BlockHeader{}
		if err := proto.Unmarshal(body, blockHeader); err == nil {
			return serializable, blockHeader, nil
		} else {
			return nil, nil, err
		}
	case MessageHeader_TRANSACTION_DATA:
		transactionData := &TransactionData{}
		if err := proto.Unmarshal(body, transactionData); err == nil {
			return serializable, transactionData, nil
		} else {
			return nil, nil, err
		}
	default:
		return nil, nil, errors.New("message header not found")
	}
}