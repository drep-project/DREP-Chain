package consensus

import (
    "BlockChainTest/network"
    "sync"
    "BlockChainTest/bean"
    "BlockChainTest/crypto"
    "math/big"
    "BlockChainTest/hash"
    "fmt"
    "math"
    "github.com/golang/protobuf/proto"
)

const (
    waiting              = 0
    setUp               = 1
    challenge            = 2
)
type Leader struct {
    members    []*network.Peer
    pubKey     *bean.Point
    state      int
    LeaderPeer *network.Peer

    commitWg sync.WaitGroup
    commitBitmap map[string] bool
    sigmaPubKey *bean.Point
    sigmaQ *bean.Point
    r []byte

    sigmaS *big.Int
    responseWg sync.WaitGroup
    responseBitmap map[string] bool

    sigs map[bean.Address][]byte
}

func NewLeader(pubKey *bean.Point, peers []*network.Peer) *Leader {
    l := &Leader{}
    l.pubKey = pubKey
    l.members = make([]*network.Peer, len(peers) - 1)
    last := 0
    for _, peer := range peers {
        if !peer.PubKey.Equal(pubKey) {
            l.members[last] = peer
            last++
        }
    }
    l.state = waiting
    l.sigmaPubKey = &bean.Point{X: []byte{0x00}, Y: []byte{0x00}}
    l.sigmaQ = &bean.Point{X: []byte{0x00}, Y: []byte{0x00}}
    l.sigmaS = new(big.Int)
    l.commitBitmap = make(map[string]bool)
    l.responseBitmap = make(map[string]bool)
    return l
}

func (l *Leader) ProcessConsensus(msg []byte) *bean.Signature {
    l.commitWg = sync.WaitGroup{}
    l.commitWg.Add(len(l.members))
    l.state = setUp
    fmt.Println("Leader is going to setup")
    l.setUp(msg, l.pubKey)
    fmt.Println("Leader wait for commit")
    l.commitWg.Wait()

    l.responseWg = sync.WaitGroup{}
    l.responseWg.Add(len(l.commitBitmap))
    l.state = challenge
    fmt.Println("Leader is going to challenge")
    l.challenge(msg)
    fmt.Println("Leader wait for response")
    l.responseWg.Wait()
    fmt.Println("Leader finish")
    return &bean.Signature{R: l.r, S: l.sigmaS.Bytes()}
}

func (l *Leader) setUp(msg []byte, pubKey *bean.Point) {
    setup := &bean.Setup{Msg: msg, PubKey: pubKey}
    fmt.Println("Leader setup ", *setup)
    network.SendMessage(l.members, setup)
}

func (l *Leader) getR(msg []byte) []byte {
    curve := crypto.GetCurve()
    r := hash.ConcatHash256(l.sigmaQ.Bytes(), l.sigmaPubKey.Bytes(), msg)
    rInt := new(big.Int).SetBytes(r)
    rInt.Mod(rInt, curve.N)
    return rInt.Bytes()
}

func (l *Leader) challenge(msg []byte)  {
    l.r = l.getR(msg)
    challenge := &bean.Challenge{SigmaPubKey: l.sigmaPubKey, SigmaQ: l.sigmaQ, R: l.r}
    fmt.Println("Leader challenge ", *challenge)
    network.SendMessage(l.members, challenge)
}

func (l *Leader) ProcessCommit(commit *bean.Commitment) {
    fmt.Println("Leader process commit ", *commit)
    if l.state != setUp {
        return
    }
    //if !store.CheckRole(node.LEADER) {
    //    return
    //}
    addr := commit.PubKey.Addr()
    if l.commitBitmap[addr] {
       return
    }
    l.commitBitmap[addr] = true
    l.commitWg.Done()
    curve := crypto.GetCurve()
    l.sigmaPubKey = curve.Add(l.sigmaPubKey, commit.PubKey)
    l.sigmaQ = curve.Add(l.sigmaQ, commit.Q)
}

func (l *Leader) ProcessResponse(response *bean.Response) {
    fmt.Println("Leader process response ", *response)
    if l.state != challenge {
        return
    }
    //if !store.CheckRole(node.LEADER) {
    //    return
    //}
    addr := response.PubKey.Addr()
    if l.responseBitmap[addr] {
       return
    }
    l.responseBitmap[addr] = true
    l.responseWg.Done()
    s := new(big.Int).SetBytes(response.S)
    l.sigmaS = l.sigmaS.Add(l.sigmaS, s)
}

func (l *Leader) Validate(sig *bean.Signature) bool {
    if len(l.responseBitmap) < len(l.commitBitmap) {
        return false
    }
    if float64(len(l.responseBitmap)) < math.Ceil(float64(len(l.members)*2.0/3.0)+1) {
        return false
    }
    challenge := &bean.Challenge{SigmaPubKey: l.sigmaPubKey, SigmaQ: l.sigmaQ, R: l.r}
    b, err := proto.Marshal(challenge)
    if err != nil {
        return false
    }
    return crypto.Verify(sig, l.sigmaPubKey, b)
}