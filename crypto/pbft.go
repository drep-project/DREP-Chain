package crypto

import (
    "math"
    "math/big"
    "BlockChainTest/common"
    "BlockChainTest/network"
)

var (
    COMMAND = []byte("Please send your ticket to me")
    ANNOUNCEMENT = []byte("Now begins the multi_signature")
)

type Leader struct {
	Command        *common.Word
	Announcement   *common.Word
	Challenge      *common.Challenge
	Object         interface{}
	Minors         []*Minor
	EnterOK        map[string] *common.Ticket
	CommitOK       map[string] *common.Commitment
	RespondOK      map[string] *common.Response
	Task           *network.Task
}

func (leader *Leader) InitCommand() {
    leader.Command = &common.Word{Msg: COMMAND}
}

func (leader *Leader) InitAnnouncement() {
    leader.Announcement = &common.Word{Msg: ANNOUNCEMENT}
}

func (leader *Leader) InitObject() {
    // TODO
}

func (leader *Leader) RequestTicket() error {

   return NewBroadcast(leader.MinorAddresses, MinorPort, leader.Command, leader.Net.BroadcastQueue).Spread()
}

func (leader *Leader) CollectTicket(ticket *Ticket, ip string) {
   if leader.EnterOK == nil {
       leader.EnterOK = make(map[string] *Ticket)
   }
   leader.EnterOK[ip] = ticket
}

func (leader *Leader) ValidateTicket() bool {
   if len(leader.EnterOK) != len(leader.MinorPubKeys) {
       return false
   }
   for ip, p := range leader.MinorPubKeys {
       if q, b := leader.EnterOK[ip]; !b {
           return false
       } else if !PointEqual(p, q.PubKey) {
           return false
       } else if !Verify(curve, q.SigOfWord, q.PubKey, leader.Command.Msg) {
           return false
       }
   }
   return true
}

func (leader *Leader) RequestCommitment() error {
   return NewBroadcast(leader.MinorAddresses, MinorPort, leader.Announcement, leader.Net.BroadcastQueue).Spread()
}

func (leader *Leader) CollectCommitment(commitment *Commitment, ip string) {
   if leader.CommitOK == nil {
       leader.CommitOK = make(map[string] *Commitment)
   }
   leader.CommitOK[ip] = commitment
}

func (leader *Leader) ProposeChallenge() error {
   msg, err := Serialize(leader.Object)
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
   r := ConcatHash(groupQ, groupPubKey, msg)
   challenge := &Challenge{GroupPubKey: groupPubKey, GroupQ: groupQ, Msg: msg, R: r}
   leader.Challenge = challenge
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
   if float64(len(leader.RespondOK)) < math.Ceil(float64(len(leader.MinorPubKeys) * 2.0 / 3.0) + 1) {
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
   sig := &Signature{R: leader.Challenge.R, S: groupS.Bytes()}
   msg, err := Serialize(leader.Object)
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

type Minor struct {
	PrvKey *PrivateKey
	K      []byte
	Net    *Network
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
   k, err := GetRandomKQ(curve)
   if err != nil {
       return err
   }
   minor.K = k
   ips := append(make([]string, 0), ip)
   commitment := &Commitment{PubKey: minor.PrvKey.PubKey, Q: curve.ScalarBaseMultiply(k)}
   return NewBroadcast(ips, LeaderPort, commitment, minor.Net.BroadcastQueue).Spread()
}

func (minor *Minor) ReturnResponse(challenge *Challenge, ip string) error {
   r := ConcatHash(challenge.GroupQ, challenge.GroupPubKey, challenge.Msg)
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