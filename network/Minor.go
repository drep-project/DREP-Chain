package network

import (
    "math/big"
    "errors"
)

type Minor struct {
    PrvKey *PrivateKey
    K      []byte
    Net    *Network
}

func NewMinor (peer *Peer) *Minor {
    /*
    if peer.AsLeader != nil || peer.AsMinor != nil {
        return errors.New("fail to setup minor, currently involved in another signing protocol")
    }
    */
    minor := &Minor{}
    //minor.PrvKey = peer.PrvKey
    //minor.Net = peer.Net
    return minor
}

func (minor *Minor) ReturnTicket(word *CommandOfWord, ip string) error {
    ips := append(make([]string, 0), ip)
    sigOfWord, err := Sign(curve, minor.PrvKey, []byte(word.Msg))
    if err != nil {
        return err
    }
    ticket := &Ticket{}
    ticket.PubKey = minor.PrvKey.PubKey
    ticket.SigOfWord = sigOfWord
    return NewBroadcast(ips, LeaderPort, ticket, minor.Net.BroadcastQueue).Spread()
}

func (minor *Minor) ReturnCommitment(signal *SignalOfStart, ip string) error {
    k, err := RandomSample(curve)
    if err != nil {
        return err
    }
    minor.K = k
    ips := append(make([]string, 0), ip)
    commitment := &Commitment{PubKey: minor.PrvKey.PubKey, Q: curve.ScalarBaseMultiply(k)}
    return NewBroadcast(ips, LeaderPort, commitment, minor.Net.BroadcastQueue).Spread()
}

func (minor *Minor) ReturnResponse(challenge *Challenge, ip string) error {
    r := ConcatEncode(challenge.GroupQ, challenge.GroupPubKey, challenge.Msg)
    r0 := new(big.Int).SetBytes(challenge.R)
    r1 := new(big.Int).SetBytes(r)
    if r0.Cmp(r1) != 0 {
        return errors.New("wrong hash value")
    }
    kInt := new(big.Int).SetBytes(minor.K)
    prvInt := new(big.Int).SetBytes(minor.PrvKey.Prv)
    s := new(big.Int).Mul(r1, prvInt)
    s.Sub(kInt, s)
    s.Mod(s, curve.N)
    response := &Response{PubKey: minor.PrvKey.PubKey, S: s.Bytes()}
    ips := append(make([]string, 0), ip)
    return NewBroadcast(ips, LeaderPort, response, minor.Net.BroadcastQueue).Spread()
}

func (minor *Minor) ListeningPort() int {
    return MinorPort
}

func (minor *Minor) ListeningRole() interface{} {
    return minor
}

func (minor *Minor) BroadcastQueue() chan *Sender {
    return minor.Net.BroadcastQueue
}

func (minor *Minor) NotificationQueue() chan *Notification {
    return minor.Net.NotificationQueue
}

func (minor *Minor) Listen() error {
    Listen(minor)
    return nil
}

func (minor *Minor) Work() error {
    Work(minor)
    return nil
}