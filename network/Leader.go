package network

import (
    "math"
    "math/big"
)

type Leader struct {
    Word            *CommandOfWord
    Signal          *SignalOfStart
    Chn             *Challenge
    Plaintext       interface{}
    RosterIPs       []string
    Roster          map[string] *Point
    EnterOK         map[string] *Ticket
    CommitOK        map[string] *Commitment
    RespondOK       map[string] *Response
    Net             *Network
}

func NewLeader(peer *Peer) *Leader {
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
    leader.Net = peer.Net
    return leader
}

func (leader *Leader) RequestTicket() error {
    return NewBroadcast(leader.RosterIPs, MinorPort, leader.Word, leader.Net.BroadcastQueue).Spread()
}

func (leader *Leader) CollectTicket(ticket *Ticket, ip string) {
    if leader.EnterOK == nil {
        leader.EnterOK = make(map[string] *Ticket)
    }
    leader.EnterOK[ip] = ticket
}

func (leader *Leader) ValidateTicket() bool {
    if len(leader.EnterOK) != len(leader.Roster) {
        return false
    }
    for ip, p := range leader.Roster {
        if q, b := leader.EnterOK[ip]; !b {
            return false
        } else if !PointEqual(p, q.PubKey) {
            return false
        } else if !Verify(curve, q.SigOfWord, q.PubKey, leader.Word.Msg) {
            return false
        }
    }
    return true
}

func (leader *Leader) RequestCommitment() error {
    return NewBroadcast(leader.RosterIPs, MinorPort, leader.Signal, leader.Net.BroadcastQueue).Spread()
}

func (leader *Leader) CollectCommitment(commitment *Commitment, ip string) {
    if leader.CommitOK == nil {
        leader.CommitOK = make(map[string] *Commitment)
    }
    leader.CommitOK[ip] = commitment
}

func (leader *Leader) ProposeChallenge() error {
    msg, err := Serialize(leader.Plaintext)
    if err != nil {
        return err
    }
    groupPubKey := &Point{}
    groupQ := &Point{}
    ips := make([]string, 0)
    for ip, commitment := range leader.CommitOK {
        ips = append(ips, ip)
        groupPubKey = curve.Add(groupPubKey, commitment.PubKey)
        groupQ = curve.Add(groupQ, commitment.Q)
    }
    r := ConcatEncode(groupQ, groupPubKey, msg)
    challenge := &Challenge{GroupPubKey: groupPubKey, GroupQ: groupQ, Msg: msg, R: r}
    leader.Chn = challenge
    return NewBroadcast(ips, MinorPort, challenge, leader.Net.BroadcastQueue).Spread()
}

func (leader *Leader) CollectResponse(response *Response, ip string) {
    if leader.RespondOK == nil {
        leader.RespondOK = make(map[string] *Response)
    }
    leader.RespondOK[ip] = response
}

func (leader *Leader) ValidateResponses() bool {
    if len(leader.RespondOK) < len(leader.CommitOK) {
        return false
    }
    if float64(len(leader.RespondOK)) < math.Ceil(float64(len(leader.Roster) * 2.0 / 3.0) + 1) {
        return false
    }
    groupPubKey := &Point{}
    groupS := new(big.Int)
    for ip, response := range leader.RespondOK {
        if q, b := leader.CommitOK[ip]; !b {
            return false
        } else if !PointEqual(response.PubKey, q.PubKey) {
            return false
        }
        groupPubKey = curve.Add(groupPubKey, response.PubKey)
        groupS.Add(groupS, new(big.Int).SetBytes(response.S))
    }
    groupS.Mod(groupS, curve.N)
    sig := &Signature{R: leader.Chn.R, S: groupS.Bytes()}
    msg, err := Serialize(leader.Plaintext)
    if err != nil {
        return false
    }
    return Verify(curve, sig, groupPubKey, msg)
}

func (leader *Leader) ListeningPort() int {
    return LeaderPort
}

func (leader *Leader) ListeningRole() interface{} {
    return leader
}

func (leader *Leader) BroadcastQueue() chan *Sender {
    return leader.Net.BroadcastQueue
}

func (leader *Leader) NotificationQueue() chan *Notification {
    return leader.Net.NotificationQueue
}

func (leader *Leader) Listen() error {
    Listen(leader)
    return nil
}

func (leader *Leader) Work() error {
    Work(leader)
    return nil
}

func GetRoster() ([]string, map[string] *Point) {
    rosterIPs := []string{local}
    roster := make(map[string] *Point)
    roster[local] = GetPrvKey().PubKey
    return rosterIPs, roster
}

func GetPlaintext() interface{} {
    return &CommandOfWord{Msg: []byte("please confirm this block")}
}