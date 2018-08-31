package network

import (
    "BlockChainTest/bean"
    "BlockChainTest/crypto"
    "github.com/golang/protobuf/proto"
    "net"
    "fmt"
)

type Task struct {
    Peer *Peer
    Msg  interface{}
}

func identifyMessage(message *Task) (int, interface{}) {
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
    case *bean.Block:
        return bean.MsgTypeBlock, msg.(*bean.Block)
    default:
        return -1, nil
    }
}

func (t *Task) Cipher() ([]byte, error) {
    serializable, err := bean.Serialize(t.Msg)
    if err != nil {
        return nil, err
    }
    //sig, err := crypto.Sign(serializable.Body)
    //if err != nil {
    //   return nil, err
    //}
    //serializable.Sig = sig
    //pubKey, err := crypto.GetPubKey()
    //if err != nil {
    //   return nil, err
    //}
    //serializable.PubKey = pubKey
    //plaintext, err := proto.Marshal(serializable)
    //if err != nil {
    //   return nil, err
    //}
    //cipher, err := crypto.Encrypt(m.Peer.PubKey, plaintext)
    //if err != nil {
    //   return nil, err
    //}
    //return cipher, nil
    serializable.Sig = &bean.Signature{R: []byte{0x00}, S: []byte{0x00}}
    serializable.PubKey = &bean.Point{X: []byte{0x00}, Y: []byte{0x00}}
    return proto.Marshal(serializable)
}

func (t *Task) Send() error {
    // If sleep 1000 here, haha
    cipher, err := t.Cipher()
    if err != nil {
        return err
    }
    addr, err := net.ResolveTCPAddr("tcp", t.Peer.ToString())
    if err != nil {
        return err
    }
    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return err
    }
    defer conn.Close()
    fmt.Println("Send msg to ",t.Peer.ToString(), cipher)
    if num, err := conn.Write(cipher); err != nil {
        fmt.Println("Send error ", err)
        return err
    } else {
        fmt.Println("Send bytes ", num)
        return nil
    }
}

func DecryptIntoMessage(cipher []byte) (*Task, error) {
    plaintext, err := crypto.Decrypt(cipher)
    if err != nil {
        return nil, err
    }
    serializable, msg, err := bean.Deserialize(plaintext)
    if err != nil {
        return nil, err
    }
    //if !crypto.Verify(serializable.Sig, serializable.PubKey, serializable.Body) {
    //   return nil, errors.New("decrypt fail")
    //}
    peer := &Peer{PubKey: serializable.PubKey}
    message := &Task{Peer: peer, Msg: msg}
    return message, nil
}