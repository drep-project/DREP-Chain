package common

import (
	"github.com/golang/protobuf/proto"
	"errors"
)

func Serialize(message interface{}) ([]byte, error) {
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
	case *Word:
		serializable.Header = MessageHeader_COMMAND_OF_WORD
	case *Ticket:
		serializable.Header = MessageHeader_TICKET
	case *Commitment:
		serializable.Header = MessageHeader_COMMITMENT
	case *Challenge:
		serializable.Header = MessageHeader_CHALLENGE
	case *Response:
		serializable.Header = MessageHeader_RESPONSE
	default:
		return nil, errors.New("bad message type")
	}
	return proto.Marshal(serializable)
}

func Deserialize(b []byte) (interface{}, error) {
	serializable := &Serializable{}
	if err := proto.Unmarshal(b, serializable); err != nil {
		return nil, err
	}
	body := serializable.GetBody()
	switch serializable.GetHeader() {
	case MessageHeader_POINT:
		point := &Point{}
		if err := proto.Unmarshal(body, point); err == nil {
			return point, nil
		} else {
			return nil, err
		}
	case MessageHeader_PRIVATE_KEY:
		prvKey := &PrivateKey{}
		if err := proto.Unmarshal(body, prvKey); err == nil {
			return prvKey, nil
		} else {
			return nil, err
		}
	case MessageHeader_SIGNATURE:
		sig := &Signature{}
		if err := proto.Unmarshal(body, sig); err == nil {
			return sig, nil
		} else {
			return nil, err
		}
	case MessageHeader_WORD:
		word := &Word{}
		if err := proto.Unmarshal(body, word); err == nil {
			return word, nil
		} else {
			return nil, err
		}
	case MessageHeader_TICKET:
		ticket := &Ticket{}
		if err := proto.Unmarshal(body, ticket); err == nil {
			return ticket, nil
		} else {
			return nil, err
		}
	case MessageHeader_COMMITMENT:
		commitment := &Commitment{}
		if err := proto.Unmarshal(body, commitment); err == nil {
			return commitment, nil
		} else {
			return nil, err
		}
	case MessageHeader_CHALLENGE:
		challenge := &Challenge{}
		if err := proto.Unmarshal(body, challenge); err == nil {
			return challenge, nil
		} else {
			return nil, err
		}
	case MessageHeader_RESPONSE:
		response := &Response{}
		if err := proto.Unmarshal(body, response); err == nil {
			return response, nil
		} else {
			return nil, err
		}
	default:
		return nil, errors.New("message header not found")
	}
}