package network

import (
    "BlockChainTest/bean"
    "BlockChainTest/log"
    "net"
    "fmt"
    "strconv"
    "BlockChainTest/mycrypto"
    "time"
    "BlockChainTest/util"
    "encoding/json"
)

type Task struct {
    PrvKey *mycrypto.PrivateKey
    Peer *bean.Peer
    Msg  interface{}
}

func (t *Task) cipher() ([]byte, error) {
    serializable, err := bean.Serialize(t.Msg)
    if err != nil {
        log.Error("there's an error during the serialize", err.Error())
        return nil, err
    }
    sig, err := mycrypto.Sign(t.PrvKey, serializable.Body)
    if err != nil {
      return nil, err
    }
    serializable.Sig = sig
    serializable.PubKey = t.PrvKey.PubKey
    return json.Marshal(serializable)
}

func (t *Task) execute() error {
    cipher, err := t.cipher()
    if err != nil {
        log.Info("error during cipher:", err)
        return &util.DataError{MyError:util.MyError{Err:err}}
    }
    d, err := time.ParseDuration("3s")
    if err != nil {
        log.Error(err.Error())
        return &util.DefaultError{}
    }
    var conn net.Conn
    for i := 0; i <= 2; i++ {
        conn, err = net.DialTimeout("tcp", t.Peer.ToString(), d)
        if err == nil {
            break
        } else {
            log.Info(fmt.Sprintf("%T %v\n", err, err))
            if ope, ok := err.(*net.OpError); ok {
                log.Info(strconv.FormatBool(ope.Timeout()), ope)
            }
            log.Info("Retry after 2s")
            time.Sleep(2 * time.Second)
        }
    }
    if err != nil {
        log.Info(fmt.Sprintf("%T %v\n", err, err))
        if ope, ok := err.(*net.OpError); ok {
            log.Info(strconv.FormatBool(ope.Timeout()), ope)
            if ope.Timeout() {
                return &util.TimeoutError{MyError:util.MyError{Err:ope}}
            } else {
                return &util.ConnectionError{MyError:util.MyError{Err:ope}}
            }
        }
    }
    defer conn.Close()
    now := time.Now()
    d2, err := time.ParseDuration("5s")
    if err != nil {
        log.Error(err.Error())
        return &util.DefaultError{}
    } else {
        conn.SetDeadline(now.Add(d2))
    }
    log.Info("Send msg to ",t.Peer.ToString(), cipher)
    if num, err := conn.Write(cipher); err != nil {
        log.Info("Send error ", err)
        return &util.TransmissionError{MyError:util.MyError{Err:err}}
    } else {
        log.Info("Send bytes ", num)
        return nil
    }
}