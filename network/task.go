package network

import (
    "BlockChainTest/bean"
    "github.com/golang/protobuf/proto"
    "BlockChainTest/log"
    "net"
    "fmt"
    "BlockChainTest/mycrypto"
    "time"
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
    //sig, err := mycrypto.Sign(serializable.Body)
    //if err != nil {
    //   return nil, err
    //}
    //serializable.Sig = sig
    //pubKey, err := mycrypto.GetPubKey()
    //if err != nil {
    //   return nil, err
    //}
    //serializable.PubKey = pubKey
    //plaintext, err := proto.Marshal(serializable)
    //if err != nil {
    //   return nil, err
    //}
    //cipher, err := mycrypto.Encrypt(m.Peer.PubKey, plaintext)
    //if err != nil {
    //   return nil, err
    //}
    //return cipher, nil
    serializable.Sig = &mycrypto.Signature{R: []byte{0x00}, S: []byte{0x00}}
    serializable.PubKey = &mycrypto.Point{X: []byte{0x00}, Y: []byte{0x00}}
    return proto.Marshal(serializable)
}

func (t *Task) execute() error {
    cipher, err := t.cipher()
    if err != nil {
        log.Println("error during cipher:", err)
        return &DataError{myError{err}}
    }
    d, err := time.ParseDuration("3s")
    if err != nil {
        fmt.Println(err)
        return &DefaultError{}
    }
    var conn net.Conn
    for i := 0; i <= 2; i++ {
        conn, err = net.DialTimeout("tcp", t.Peer.ToString(), d)
        if err == nil {
            break
        } else {
            fmt.Printf("%T %v\n", err, err)
            if ope, ok := err.(*net.OpError); ok {
                fmt.Println(ope.Timeout(), ope)
            }
            fmt.Println("Retry after 2s")
            time.Sleep(2 * time.Second)
        }
    }
    if err != nil {
        fmt.Printf("%T %v\n", err, err)
        if ope, ok := err.(*net.OpError); ok {
            fmt.Println(ope.Timeout(), ope)
            if ope.Timeout() {
                return &TimeoutError{myError{ope}}
            } else {
                return &ConnectionError{myError{ope}}
            }
        }
    }
    defer conn.Close()
    now := time.Now()
    d2, err := time.ParseDuration("5s")
    if err != nil {
        fmt.Println(err)
        return &DefaultError{}
    } else {
        conn.SetDeadline(now.Add(d2))
    }
    log.Println("Send msg to ",t.Peer.ToString(), cipher)
    if num, err := conn.Write(cipher); err != nil {
        log.Println("Send error ", err)
        return &TransmissionError{myError{err}}
    } else {
        log.Println("Send bytes ", num)
        return nil
    }
}