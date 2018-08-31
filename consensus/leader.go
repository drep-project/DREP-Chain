package consensus

import (
    "BlockChainTest/network"
    "sync"
    "BlockChainTest/bean"
    "BlockChainTest/crypto"
    "math/big"
    "math"
    "BlockChainTest/log"
)

const (
    waiting              = 0
    setUp               = 1
    challenge            = 2
)
type Leader struct {
    members    []*network.Peer
    pubKey     *crypto.Point
    state      int
    LeaderPeer *network.Peer

    commitWg sync.WaitGroup
    commitBitmap []byte
    sigmaPubKey *crypto.Point
    sigmaQ *crypto.Point
    r []byte

    sigmaS *big.Int
    responseWg sync.WaitGroup
    responseBitmap []byte

    //sigs map[bean.Address][]byte
}

func NewLeader(pubKey *crypto.Point, members []*network.Peer) *Leader {
    l := &Leader{}
    l.pubKey = pubKey
    l.members = make([]*network.Peer, len(members) - 1)
    last := -1
    for _, v := range members {
        if v.PubKey.Equal(pubKey) {
            continue
        }
        last++
        l.members[last] = v
        //l.members = append(l.members, v)
    }
    l.state = waiting
    l.sigmaPubKey = &crypto.Point{X: []byte{0x00}, Y: []byte{0x00}}
    l.sigmaQ = &crypto.Point{X: []byte{0x00}, Y: []byte{0x00}}
    l.sigmaS = new(big.Int)
    len := len(members)
    l.commitBitmap = make([]byte, len)
    l.responseBitmap = make([]byte, len)
    l.commitWg = sync.WaitGroup{}
    l.commitWg.Add(len - 1)
    return l
}

func (l *Leader) ProcessConsensus(msg []byte) (*crypto.Signature, []byte) {

    l.state = setUp
    log.Println("Leader is going to setup")
    l.setUp(msg, l.pubKey)
    log.Println("Leader wait for commit")
    l.commitWg.Wait()

    l.responseWg = sync.WaitGroup{}
    l.responseWg.Add(len(l.commitBitmap) - 1)
    l.state = challenge
    log.Println("Leader is going to challenge")
    l.challenge(msg)
    log.Println("Leader wait for response")
    l.responseWg.Wait()
    log.Println("Leader finish")
    sig := &crypto.Signature{R: l.r, S: l.sigmaS.Bytes()}
    valid := l.Validate(sig, msg)
    log.Println("valid? ", valid)
    return sig, l.responseBitmap
}

func (l *Leader) setUp(msg []byte, pubKey *crypto.Point) {
    setup := &bean.Setup{Msg: msg, PubKey: pubKey}
    log.Println("Leader setup ", *setup)
    network.SendMessage(l.members, setup)
}

func (l *Leader) getR(msg []byte) []byte {
    curve := crypto.GetCurve()
    r := crypto.ConcatHash256(l.sigmaQ.Bytes(), l.sigmaPubKey.Bytes(), msg)
    rInt := new(big.Int).SetBytes(r)
    rInt.Mod(rInt, curve.N)
    return rInt.Bytes()
}

func (l *Leader) challenge(msg []byte)  {
    l.r = l.getR(msg)
    challenge := &bean.Challenge{SigmaPubKey: l.sigmaPubKey, SigmaQ: l.sigmaQ, R: l.r}
    log.Println("Leader challenge ", *challenge)
    network.SendMessage(l.members, challenge)
}

func isLegalIndex(index int, bitmap []byte) bool {
    return index >=0 && index <= len(bitmap) && bitmap[index] != 1
}

func (l *Leader) getMinerIndex(p *crypto.Point) int {
    if l.pubKey.Equal(p) {
        return -1
    }
    for i, v := range l.members {
        if v.PubKey.Equal(p) {
            return i
        }
    }
    return -1
}

func (l *Leader) ProcessCommit(commit *bean.Commitment) {
    log.Println("Leader process commit ", *commit)
    if l.state != setUp {
        return
    }
    //if !store.CheckRole(node.LEADER) {
    //    return
    //}
    index := l.getMinerIndex(commit.PubKey)
    if !isLegalIndex(index, l.commitBitmap) {
       return
    }
    l.commitBitmap[index] = 1
    //l.commitWg.Done()
    curve := crypto.GetCurve()
    l.sigmaPubKey = curve.Add(l.sigmaPubKey, commit.PubKey)
    l.sigmaQ = curve.Add(l.sigmaQ, commit.Q)
    l.commitWg.Done()
}

func (l *Leader) ProcessResponse(response *bean.Response) {
    log.Println("Leader process response ", *response)
    if l.state != challenge {
        return
    }
    //if !store.CheckRole(node.LEADER) {
    //    return
    //}
    index := l.getMinerIndex(response.PubKey)
    if !isLegalIndex(index, l.responseBitmap) {
       return
    }
    l.responseBitmap[index] = 1
    l.responseWg.Done()
    s := new(big.Int).SetBytes(response.S)
    l.sigmaS = l.sigmaS.Add(l.sigmaS, s)
    l.sigmaS.Mod(l.sigmaS, crypto.GetCurve().N)
}

func (l *Leader) Validate(sig *crypto.Signature, msg []byte) bool {
    if len(l.responseBitmap) < len(l.commitBitmap) {
        return false
    }
    if float64(len(l.responseBitmap)) < math.Ceil(float64(len(l.members)*2.0/3.0)+1) {
        return false
    }
    return crypto.Verify(sig, l.sigmaPubKey, msg)
}