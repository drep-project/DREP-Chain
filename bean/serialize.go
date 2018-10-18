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
	//case *Point:
	//	serializable.Header = MessageHeader_POINT
	//case *PrivateKey:
	//	serializable.Header = MessageHeader_PRIVATE_KEY
	//case *Signature:
	//	serializable.Header = MessageHeader_SIGNATURE
	case *Setup:
		serializable.Header = MsgTypeSetUp
	case *Commitment:
		serializable.Header = MsgTypeCommitment
	case *Challenge:
		serializable.Header = MsgTypeChallenge
	case *Response:
		serializable.Header = MsgTypeResponse
	case *BlockHeader:
		serializable.Header = MsgTypeBlockHeader
	case *Transaction:
		serializable.Header = MsgTypeTransaction
	case *PeerInfo:
		serializable.Header = MsgTypeNewPeer
	case *PeerInfoList:
		serializable.Header = MsgTypePeerList
	case *Block:
		serializable.Header = MsgTypeBlock
	case *BlockReq:
		serializable.Header = MsgTypeBlockReq
	case *BlockResp:
		serializable.Header = MsgTypeBlockResp
	case *Ping:
		serializable.Header = MsgTypePing
	case *Pong:
		serializable.Header = MsgTypePong
	case *OfflinePeers:
		serializable.Header = MsgTypeOfflinePeers
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
	//case MessageHeader_POINT:
	//	point := &Point{}
	//	if err := proto.Unmarshal(body, point); err == nil {
	//		return serializable, point, nil
	//	} else {
	//		return nil, nil, err
	//	}
	//case MessageHeader_PRIVATE_KEY:
	//	prvKey := &PrivateKey{}
	//	if err := proto.Unmarshal(body, prvKey); err == nil {
	//		return serializable, prvKey, nil
	//	} else {
	//		return nil, nil, err
	//	}
	//case MessageHeader_SIGNATURE:
	//	sig := &Signature{}
	//	if err := proto.Unmarshal(body, sig); err == nil {
	//		return serializable, sig, nil
	//	} else {
	//		return nil, nil, err
	//	}
	case MsgTypeSetUp:
		setup := &Setup{}
		if err := proto.Unmarshal(body, setup); err == nil {
			return serializable, setup, nil
		} else {
			return nil, nil, err
		}
	case MsgTypeCommitment:
		commitment := &Commitment{}
		if err := proto.Unmarshal(body, commitment); err == nil {
			return serializable, commitment, nil
		} else {
			return nil, nil, err
		}
	case MsgTypeChallenge:
		challenge := &Challenge{}
		if err := proto.Unmarshal(body, challenge); err == nil {
			return serializable, challenge, nil
		} else {
			return nil, nil, err
		}
	case MsgTypeResponse:
		response := &Response{}
		if err := proto.Unmarshal(body, response); err == nil {
			return serializable, response, nil
		} else {
			return nil, nil, err
		}
	case MsgTypeBlockHeader:
		blockHeader := &BlockHeader{}
		if err := proto.Unmarshal(body, blockHeader); err == nil {
			return serializable, blockHeader, nil
		} else {
			return nil, nil, err
		}
	case MsgTypeTransaction:
		transaction := &Transaction{}
		if err := proto.Unmarshal(body, transaction); err == nil {
			return serializable, transaction, nil
		} else {
			return nil, nil, err
		}
	case MsgTypeBlock:
		block := &Block{}
		if err := proto.Unmarshal(body, block); err == nil {
			return serializable, block, nil
		} else {
			return nil, nil, err
		}
	case MsgTypeNewPeer:
		peer := &PeerInfo{}
		if err := proto.Unmarshal(body, peer); err == nil {
			return serializable, peer, nil
		} else {
			return nil, nil, err
		}
	case MsgTypePeerList:
		list := &PeerInfoList{}
		if err := proto.Unmarshal(body, list); err == nil {
			return serializable, list, nil
		} else {
			return nil, nil, err
		}
    case MsgTypeBlockReq:
        blockReq := &BlockReq{}
        if err := proto.Unmarshal(body, blockReq); err == nil {
            return serializable, blockReq, nil
        } else {
            return nil, nil, err
        }
    case MsgTypeBlockResp:
        blockResp := &BlockResp{}
        if err := proto.Unmarshal(body, blockResp); err == nil {
            return serializable, blockResp, nil
        } else {
            return nil, nil, err
        }
	case MsgTypePing:
		ping := &Ping{}
		if err := proto.Unmarshal(body, ping); err == nil {
			return serializable, ping, nil
		} else {
			return nil, nil, err
		}
	case MsgTypePong:
		pong := &Pong{}
		if err := proto.Unmarshal(body, pong); err == nil {
			return serializable, pong, nil
		} else {
			return nil, nil, err
		}
	case MsgTypeOfflinePeers:
		peers := &OfflinePeers{}
		if err := proto.Unmarshal(body, peers); err == nil {
			return serializable, peers, nil
		} else {
			return nil, nil, err
		}
	default:
		return nil, nil, errors.New("message header not found")
	}
}

func Marshal(msg interface{}) ([]byte, error) {
	switch msg.(type) {
	case *BlockHeader:
		return proto.Marshal(msg.(*BlockHeader))
	case *TransactionData:
		return proto.Marshal(msg.(*TransactionData))
	default:
		return nil, errors.New("bad message type")
	}
}

