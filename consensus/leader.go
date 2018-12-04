package consensus

import (
    "BlockChainTest/network"
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "math/big"
    "math"
    "BlockChainTest/log"
    "time"
    "BlockChainTest/pool"
    "BlockChainTest/util"
    "BlockChainTest/consensus/consmsg"
)

type Leader struct {
    members    []*bean.Peer

    commitBitmap []byte
    sigmaPubKey *mycrypto.Point
    sigmaQ *mycrypto.Point
    r []byte

    sigmaS *big.Int
    responseBitmap []byte

}

func NewLeader(pubKey *mycrypto.Point, members []*bean.Peer) *Leader {
    l := &Leader{}
    l.members = make([]*bean.Peer, len(members) - 1)
    last := -1
    for _, v := range members {
        if v.PubKey.Equal(pubKey) {
            continue
        }
        last++
        l.members[last] = v
    }
    l.sigmaPubKey = &mycrypto.Point{X: []byte{0x00}, Y: []byte{0x00}}
    l.sigmaQ = &mycrypto.Point{X: []byte{0x00}, Y: []byte{0x00}}
    l.sigmaS = new(big.Int)
    length := len(members)
    l.commitBitmap = make([]byte, length)
    l.responseBitmap = make([]byte, length)
    return l
}

func (l *Leader) GetMembers() []*bean.Peer {
    return l.members
}

func (l *Leader) ProcessConsensus(msg []byte) (error, *mycrypto.Signature, []byte) {
    log.Trace("Leader is going to setup")
    ps := l.setUp(msg)
    if len(ps) == 0 {
        //return &util.OfflineError{}, nil, nil
        log.Trace("It seems that you are solo.")
    }
    log.Trace("Leader wait for commit")
    if !l.waitForCommit(ps) {
        return &util.ConnectionError{}, nil, nil
    }
    log.Trace("Leader is going to challenge")
    ps = l.challenge(msg)
    log.Trace("Leader wait for response")
    l.waitForResponse(ps)
    log.Trace("Leader finish")
    sig := &mycrypto.Signature{R: l.r, S: l.sigmaS.Bytes()}
    valid := l.Validate(sig, msg)
    log.Trace("valid? ","valid", valid)
    if !valid {
        //return &util.ConnectionError{}, nil, nil
        log.Error("Error!!!!!!! If you are solo, ignore this.")
    }
    return nil, sig, l.responseBitmap
}

func (l *Leader) setUp(msg []byte) []*bean.Peer {
    setup := &bean.Setup{Msg: msg}
    log.Trace("Leader setup ", "msg", *setup)
    s, _ := network.SendMessage(l.members, setup)
    return s
}

func (l *Leader) waitForCommit(peers []*bean.Peer) bool {
    memberNum := len(peers)
    //r := make([]bool, memberNum)
    log.Trace("waitForCommit 1")
    commits := pool.ObtainMsg(memberNum, func(msg interface{}) bool {
        if m, ok := msg.(*consmsg.CommitmentMsg); ok {
            if !contains(m.Peer.PubKey, peers) {
                return false
            }
            index := l.getMinerIndex(m.Peer.PubKey)
            if !isLegalIndex(index, l.commitBitmap) {
                return false
            }
            l.commitBitmap[index] = 1
            return true
        } else {
            return false
        }
    }, 5 * time.Second)
    log.Trace("waitForCommit 2")
    if len(commits) + 1 < memberNum * 3 / 2 {
        log.Trace("waitForCommit", "Commits", len(commits),"memberNum", memberNum)
        return false
    }
    log.Trace("waitForCommit 3")
    curve := mycrypto.GetCurve()
    for _, c := range commits {
        if commit, ok := c.(*consmsg.CommitmentMsg); ok {
            log.Trace("waitForCommit 4")
            l.sigmaPubKey = curve.Add(l.sigmaPubKey, commit.Peer.PubKey)
            l.sigmaQ = curve.Add(l.sigmaQ, commit.Msg.Q)
        }
        log.Trace("waitForCommit 5")
    }
    log.Trace("waitForCommit 6")

    return true
}

func (l *Leader) waitForResponse(peers []*bean.Peer)  {
    responses := pool.ObtainMsg(len(l.members), func(msg interface{}) bool {
        if m, ok := msg.(*consmsg.ResponseMsg); ok {
            if !contains(m.Peer.PubKey, peers) {
                return false
            }
            index := l.getMinerIndex(m.Peer.PubKey)
            if !isLegalIndex(index, l.responseBitmap) {
                return false
            }
            l.responseBitmap[index] = 1
            return true
        } else {
            return false
        }
    }, 5 * time.Second)
    for _, r := range responses {
        if response, ok := r.(*consmsg.ResponseMsg); ok {
            s := new(big.Int).SetBytes(response.Msg.S)
            l.sigmaS = l.sigmaS.Add(l.sigmaS, s)
            l.sigmaS.Mod(l.sigmaS, mycrypto.GetCurve().N)
        }
    }

}
func (l *Leader) getR(msg []byte) []byte {
    curve := mycrypto.GetCurve()
    r := mycrypto.ConcatHash256(l.sigmaQ.Bytes(), l.sigmaPubKey.Bytes(), msg)
    rInt := new(big.Int).SetBytes(r)
    rInt.Mod(rInt, curve.N)
    return rInt.Bytes()
}

func (l *Leader) challenge(msg []byte) []*bean.Peer {
    l.r = l.getR(msg)
    challenge := &bean.Challenge{SigmaPubKey: l.sigmaPubKey, SigmaQ: l.sigmaQ, R: l.r}
    log.Trace("Leader challenge ", *challenge)
    ps := make([]*bean.Peer, 0)
    for i, b := range l.commitBitmap {
        if b == 1 {
            ps = append(ps, l.members[i])
        }
    }
    sp, _ := network.SendMessage(ps, challenge)
    return sp
}

func isLegalIndex(index int, bitmap []byte) bool {
    return index >=0 && index <= len(bitmap) && bitmap[index] != 1
}

func (l *Leader) getMinerIndex(p *mycrypto.Point) int {
    // TODO if it is itself
    for i, v := range l.members {
        if v.PubKey.Equal(p) {
            return i
        }
    }
    return -1
}

func (l *Leader) Validate(sig *mycrypto.Signature, msg []byte) bool {
    log.Trace("验证", "ResponseBitmap",l.responseBitmap,"CommitBitmap" ,l.commitBitmap)
    if len(l.responseBitmap) < len(l.commitBitmap) {
        log.Trace("Validate 1")
        return false
    }
    if float64(len(l.responseBitmap)) < math.Ceil(float64(len(l.members)*2.0/3.0)+1) {
        log.Trace("Validate 2")
        return false
    }
    return mycrypto.Verify(sig, l.sigmaPubKey, msg)
}

func contains(pk *mycrypto.Point, peers []*bean.Peer) bool {
    for _, p := range peers {
        if pk.Equal(p.PubKey) {
            return true
        }
    }
    return false
}