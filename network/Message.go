package network

import (
    "BlockChainTest/bean"
    "BlockChainTest/crypto"
    "github.com/golang/protobuf/proto"
    "net"
    "errors"
    "fmt"
)

type Message struct {
    Peer *Peer
    Msg        interface{}
}

func (m *Message) Cipher() ([]byte, error) {
    serializable, err := bean.Serialize(m.Msg)
    if err != nil {
        return nil, err
    }
    sig, err := crypto.Sign(serializable.Body)
    if err != nil {
        return nil, err
    }
    serializable.Sig = sig
    pubKey, err := crypto.GetPubKey()
    if err != nil {
        return nil, err
    }
    serializable.PubKey = pubKey
    plaintext, err := proto.Marshal(serializable)
    if err != nil {
        return nil, err
    }
    cipher, err := crypto.Encrypt(m.Peer.PubKey, plaintext)
    if err != nil {
        return nil, err
    }
    return cipher, nil
}

func (m *Message) Send() error {
    cipher, err := m.Cipher()
    if err != nil {
        return err
    }
    addr, err := net.ResolveTCPAddr("tcp", m.Peer.ToString())
    if err != nil {
        return err
    }
    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return err
    }
    defer conn.Close()
    fmt.Println("Send msg:", cipher)
    if _, err := conn.Write(cipher); err != nil {
        return err
    }
    return nil
}

func DecryptIntoMessage(cipher []byte) (*Message, error) {
    plaintext, err := crypto.Decrypt(cipher)
    if err != nil {
        return nil, err
    }
    serializable, msg, err := bean.Deserialize(plaintext)
    if err != nil {
        return nil, err
    }
    if !crypto.Verify(serializable.Sig, serializable.PubKey, serializable.Body) {
        return nil, errors.New("decrypt fail")
    }
    peer := &Peer{PubKey: serializable.PubKey}
    message := &Message{Peer: peer, Msg: msg}
    return message, nil
}

func identifyMessage(message *Message) (int, interface{}) {
    msg := message.Msg
    switch msg.(type) {
    case *bean.Setup:
        return bean.MsgTypeSetUp, msg.(*bean.Setup)
    case *bean.Commitment:
        return bean.MsgTypeCommitment, msg.(*bean.Commitment)
    case *bean.Challenge:
        return bean.MsgTypeChallenge, msg.(*bean.Challenge)
    case *bean.Response:
        return bean.MsgTypeResponse, msg.(*bean.Response)
    default:
        return -1, nil
    }
}