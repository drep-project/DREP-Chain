package network

import (
    "strconv"
)

type Peer struct {
    IP       string
    Port     int
    msg      interface{}
}

//func (sender *Sender) Send() error {
//    return SendMessage(sender)
//}

func (peer *Peer) Address() string{
    return peer.IP + ":" + strconv.Itoa(peer.Port)
}



func (peer *Peer) initLeader() *Leader {
    /*
    if peer.AsLeader != nil || peer.AsMinor != nil {
        return errors.New("fail to setup leader, currently involved in another signing protocol")
    }
    */
    leader := &Leader{}
    leader.Word = &CommandOfWord{Msg: []byte("please send your ticket to me")}
    leader.Signal = &SignalOfStart{Mark: 1}
    leader.Plaintext = GetPlaintext()
    leader.RosterIPs, leader.Roster = GetRoster()
    leader.EnterOK = make(map[string] *Ticket)
    leader.CommitOK = make(map[string] *Commitment)
    leader.RespondOK = make(map[string] *Response)
    //leader.Net = peer.Net
    return leader
}

func (peer *Peer) initMinor() *Minor {
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
