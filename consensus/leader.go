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
    "BlockChainTest/store"
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
    commitBitmap []byte
    sigmaPubKey *bean.Point
    sigmaQ *bean.Point
    r []byte

    sigmaS *big.Int
    responseWg sync.WaitGroup
    responseBitmap []byte

    //sigs map[bean.Address][]byte
}

func NewLeader(pubKey *bean.Point, peers []*network.Peer) *Leader {
    l := &Leader{}
    l.pubKey = pubKey
    l.members = make([]*network.Peer, len(peers) - 1)
    l.members = peers
    l.state = waiting
    l.sigmaPubKey = &bean.Point{X: []byte{0x00}, Y: []byte{0x00}}
    l.sigmaQ = &bean.Point{X: []byte{0x00}, Y: []byte{0x00}}
    l.sigmaS = new(big.Int)
    len := len(l.members)
    l.commitBitmap = make([]byte, len)
    l.responseBitmap = make([]byte, len)
    return l
}

func (l *Leader) ProcessConsensus(msg []byte) (*bean.Signature, []byte) {
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
    sig := &bean.Signature{R: l.r, S: l.sigmaS.Bytes()}
    valid := l.Validate(sig, msg)
    fmt.Println("valid? ", valid)
    return sig, l.responseBitmap
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

func isLegalIndex(index int, bitmap []byte) bool {
    return index >=0 && index <= len(bitmap) && bitmap[index] != 1
}
func (l *Leader) ProcessCommit(commit *bean.Commitment) {
    fmt.Println("Leader process commit ", *commit)
    if l.state != setUp {
        return
    }
    //if !store.CheckRole(node.LEADER) {
    //    return
    //}
    index := store.GetMinerIndex(commit.PubKey)
    if !isLegalIndex(index, l.commitBitmap) {
       return
    }
    l.commitBitmap[index] = 1
    l.commitWg.Done()
    curve := crypto.GetCurve()
    l.sigmaPubKey = curve.Add(l.sigmaPubKey, commit.PubKey)
    l.sigmaQ = curve.Add(l.sigmaQ, commit.Q)
    l.commitWg.Done()
}

func (l *Leader) ProcessResponse(response *bean.Response) {
    fmt.Println("Leader process response ", *response)
    if l.state != challenge {
        return
    }
    //if !store.CheckRole(node.LEADER) {
    //    return
    //}
    index := store.GetMinerIndex(response.PubKey)
    if !isLegalIndex(index, l.responseBitmap) {
       return
    }
    l.responseBitmap[index] = 1
    l.responseWg.Done()
    s := new(big.Int).SetBytes(response.S)
    l.sigmaS = l.sigmaS.Add(l.sigmaS, s)
    l.sigmaS.Mod(l.sigmaS, crypto.GetCurve().N)
}

func (l *Leader) Validate(sig *bean.Signature, msg []byte) bool {
    if len(l.responseBitmap) < len(l.commitBitmap) {
        return false
    }
    if float64(len(l.responseBitmap)) < math.Ceil(float64(len(l.members)*2.0/3.0)+1) {
        return false
    }
    return crypto.Verify(sig, l.sigmaPubKey, msg)
}