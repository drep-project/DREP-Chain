package network

import (
    "BlockChainTest/bean"
    "github.com/golang/protobuf/proto"
    "BlockChainTest/log"
    "net"
    "fmt"
)

type Task struct {
    Peer *Peer
    Msg  interface{}
}

func (t *Task) cipher() ([]byte, error) {
    serializable, err := bean.Serialize(t.Msg)
    if err != nil {
        fmt.Println("there's an error during the serialize", err)
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

func (t *Task) execute() error {
    // If sleep 1000 here, hahax
    cipher, err := t.cipher()

    if err != nil {
        log.Println("error during cipher:", err)
        return err
    }
    addr, err := net.ResolveTCPAddr("tcp", t.Peer.ToString())
    if err != nil {
        return err
    }
    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        log.Println("error during dail:", err)
        return err
    }
    defer conn.Close()
    log.Println("Send msg to ",t.Peer.ToString(), cipher)
    if num, err := conn.Write(cipher); err != nil {
        log.Println("Send error ", err)
        return err
    } else {
        log.Println("Send bytes ", num)
        return nil
    }
}