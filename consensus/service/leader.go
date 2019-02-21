package service

import (
    "errors"
    "fmt"
    "math"
    "math/big"
    "sync"
    "time"

    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "github.com/drep-project/drep-chain/log"
    consensusTypes "github.com/drep-project/drep-chain/consensus/types"
    p2pService "github.com/drep-project/drep-chain/network/service"
    p2pTypes "github.com/drep-project/drep-chain/network/types"
)

const (
    INIT = iota
    WAIT_SETUP
    WAIT_SETUP_TIMEOUT
    WAIT_COMMIT
    WAIT_COMMITT_IMEOUT
    WAIT_CHALLENGE
    WAIT_CHALLENGE_TIMEOUT
    WAIT_RESPONSE
    WAIT_RESPONSE_TIMEOUT
    COMPLETED
    ERROR
)

type Leader struct {
    members    []*p2pTypes.Peer
    pubkey *secp256k1.PublicKey
    p2pServer *p2pService.P2pService

    commitBitmap []byte
    sigmaPubKey []*secp256k1.PublicKey
    sigmaCommitPubkey []*secp256k1.PublicKey
    r []byte

    sigmaS *schnorr.Signature
    responseBitmap []byte
    syncLock sync.Mutex

    waitTime time.Duration

    currentHeight int64
    minMember int
    currentState int
    stateLock sync.RWMutex
    cancelWaitCommit chan struct{}
    cancelWaitChallenge chan struct{}

    quitRound chan struct{}
}

func NewLeader(pubKey *secp256k1.PublicKey, quitRound chan struct{}, p2pServer *p2pService.P2pService) *Leader {
    l := &Leader{}
    l.waitTime = 10 * time.Second
    l.pubkey = pubKey
    l.p2pServer = p2pServer
    l.quitRound = quitRound

    l.Reset()
    return l
}

func (l *Leader) UpdateStatus(members []*p2pTypes.Peer, curMiner int, minMember int, curHeight int64){
    l.members = make([]*p2pTypes.Peer, len(members) - 1)
    l.minMember = minMember //*2/3
    l.currentHeight = curHeight

    last := -1
    for _, v := range members {
        if  v == nil {
            continue
        }
        last++
        l.members[last] = v
    }
    l.Reset()
}

func (l *Leader) Reset(){
    l.sigmaPubKey = nil
    l.sigmaCommitPubkey = nil
    fmt.Println("rrrrrrrrrrrrrrrrrrrrrrrrrrrrrr")
    l.sigmaS = nil
    length := len(l.members)
    l.commitBitmap = make([]byte, length)
    l.responseBitmap = make([]byte, length)

    l.cancelWaitCommit = make(chan struct{},1)
    l.cancelWaitChallenge = make(chan struct{},1)
    l.setState(INIT)
}


func (l *Leader) ProcessConsensus(msg []byte) (error, *secp256k1.Signature, []byte) {
    l.setState(INIT)
    //TODO  Waiting here is just to reduce the chance of message staggering. This waiting time should be minimized or even eliminated.
    time.AfterFunc(time.Second, func() {
        l.setUp(msg)
    })
    if !l.waitForCommit() {
        //send reason and reset
        l.fail("waitForCommit fail")
        return errors.New("waitForCommit fail"), nil, nil
    }

    l.challenge(msg)
    if !l.waitForResponse() {
        //send reason and reset
        l.fail("waitForResponse fail")
        return errors.New("waitForResponse fail"), nil, nil
    }
    log.Debug("response complete")

    if l.sigmaS == nil {
        return errors.New("signature not valid"), nil,nil
    }
    valid := l.Validate(msg, l.sigmaS.R, l.sigmaS.S)
    fmt.Println(l.sigmaS)
    log.Debug("vaidate result","VALID", valid)
    if !valid&&len(l.members)>0 { //solo
        //return &util.ConnectionError{}, nil, nil
       return errors.New("signature not valid"), nil,nil
    }
    return nil, &secp256k1.Signature{R : l.sigmaS.R, S : l.sigmaS.S}, l.responseBitmap
}

func (l *Leader) setUp(msg []byte) {
    setup := &consensusTypes.Setup{ Msg: msg}
    setup.Height = l.currentHeight
    for _, member := range l.members {
        log.Debug("leader sent setup message", "IP", member.GetAddr(), "Height", setup.Height)
        l.p2pServer.SendAsync(member, setup)
    }
}

func (l *Leader) OnCommit(peer *p2pTypes.Peer, commit *consensusTypes.Commitment) {
    if l.getState() != WAIT_COMMIT {
        return
    }
    if l.currentHeight != commit.Height {
        return
    }

    index := l.getMinerIndex(peer.PubKey)
    if !hasMarked(index, l.commitBitmap) {
        return
    }
    l.syncLock.Lock()
    defer l.syncLock.Unlock()

    if l.sigmaPubKey == nil {
        l.sigmaPubKey = []*secp256k1.PublicKey{ peer.PubKey }
    } else {
        l.sigmaPubKey = append(l.sigmaPubKey,  peer.PubKey)
    }

    if l.sigmaCommitPubkey == nil {
        l.sigmaCommitPubkey = []*secp256k1.PublicKey{ commit.Q }
    }else{
        l.sigmaCommitPubkey = append(l.sigmaCommitPubkey, commit.Q)
    }

    l.commitBitmap[index] = 1
    commitNum := l.getCommitNum()
    if commitNum >= l.minMember  {
        log.Debug("OnCommit finish", "commitNum", commitNum, "members", len(l.members))
        select {
        case l.cancelWaitCommit <- struct{}{}:
        default:
        }
    }
}

func (l *Leader) waitForCommit() bool {
    l.setState(WAIT_COMMIT)
    for {
        select {
        case <-time.After(l.waitTime):
            commitNum := l.getCommitNum()
            log.Debug("waitForCommit  finish", "commitNum", commitNum, "members", len(l.members))
            if commitNum >= l.minMember  {
                return true
            }
            l.setState(WAIT_COMMITT_IMEOUT)
            return false
        case <-l.cancelWaitCommit:
            return true
        }
    }
}

func (l *Leader) OnResponse(peer *p2pTypes.Peer, response *consensusTypes.Response) {
    if l.getState() != WAIT_RESPONSE {
        return
    }
    if l.currentHeight != response.Height {
        return
    }

    index := l.getMinerIndex(peer.PubKey)
    if !hasMarked(index, l.responseBitmap) {
        return
    }

    l.syncLock.Lock()
    defer l.syncLock.Unlock()
    /*
    s := new(big.Int).SetBytes(response.S)
    l.sigmaS = l.sigmaS.Add(l.sigmaS, s)
    l.sigmaS.Mod(l.sigmaS, secp256k1.S256().N)
    */
    sig, err := schnorr.ParseSignature(response.S)
    if err != nil {
        return
    }

    if l.sigmaS == nil {
        l.sigmaS = sig
    }else{
        l.sigmaS, err = schnorr.CombineSigs(secp256k1.S256(),[]*schnorr.Signature{l.sigmaS, sig })
        if err != nil {
            log.Debug("schnorr CombineSigs error", "reason", err)
            return
        }else {
        }
    }

    l.responseBitmap[index] = 1

    responseNum := l.getResponseNum()
    if responseNum >= l.minMember{
        l.setState(COMPLETED)
        log.Debug("OnResponse finish", "responseNum", responseNum, "members", len(l.members))
        select {
        case  l.cancelWaitChallenge <- struct{}{}:
        default:
            return
        }
    }
}

func (l *Leader) challenge(msg []byte) {
   // l.r = l.getR(msg)
   // l.r = sha3.ConcatHash256(l.sigmaCommitPubkey.Serialize(), l.sigmaPubKey.Serialize(), msg)
    for _, member := range l.members {
        memIndex := 0
        sigmaPubKeys := []*secp256k1.PublicKey{}
        for index, pubkey := range  l.sigmaPubKey {
            if !pubkey.IsEqual(member.PubKey) {
                sigmaPubKeys = append(sigmaPubKeys, pubkey)
            }else{
                memIndex = index
            }
        }
        sigmaPubKey := schnorr.CombinePubkeys(sigmaPubKeys)

        commitPubkeys := []*secp256k1.PublicKey{}
        for index, pubkey := range  l.sigmaCommitPubkey {
            if memIndex != index {
                commitPubkeys = append(commitPubkeys, pubkey)
            }
        }
        commitPubkey := schnorr.CombinePubkeys(commitPubkeys)

        l.r = sha3.ConcatHash256(sigmaPubKey.Serialize(), commitPubkey.Serialize(), sha3.Hash256(msg))
        challenge := &consensusTypes.Challenge{Height:l.currentHeight, SigmaPubKey: sigmaPubKey, SigmaQ: commitPubkey, R: l.r}
        log.Debug("leader sent hallenge message", "IP", member.GetAddr(), "Height", l.currentHeight)
        l.p2pServer.SendAsync(member,challenge)
    }
}

func (l *Leader) fail(msg string){
CANCEL:
    for{
        select {
        case l.cancelWaitChallenge <- struct{}{}:
        case l.cancelWaitCommit <- struct{}{}:
        default:
            break CANCEL
        }
    }
    failMsg := &consensusTypes.Fail{Reason: msg}
    failMsg.Height = l.currentHeight
    for _, member := range l.members {
        l.p2pServer.SendAsync(member, failMsg)
    }
}

func (l *Leader) waitForResponse() bool {
    l.setState(WAIT_RESPONSE)
    for {
        select {
        case <-time.After(l.waitTime):
            responseNum := l.getResponseNum()
            log.Debug("waitForResponse finish", "responseNum", responseNum, "members", len(l.members))
            if responseNum  >= l.minMember{
                l.setState(COMPLETED)
                return true
            }
            l.setState(WAIT_RESPONSE_TIMEOUT)
            return false
        case <-l.cancelWaitChallenge:
            return true
        }
    }
}

func (l *Leader) getR(msg []byte) []byte {
    /*
    curve := secp256k1.S256()
    r := sha3.ConcatHash256(l.sigmaCommitPubkey.Serialize(), l.sigmaPubKey.Serialize(), msg)
    rInt := new(big.Int).SetBytes(r)
    rInt.Mod(rInt, curve.Params().N)
    return rInt.Bytes()
    */
    return []byte{}
}

func hasMarked(index int, bitmap []byte) bool {
    return index >=0 && index <= len(bitmap) && bitmap[index] != 1
}

func (l *Leader) getMinerIndex(p *secp256k1.PublicKey) int {
    // TODO if it is itself
    for i, v := range l.members {
        if v.PubKey.IsEqual(p) {
            return i
        }
    }
    return -1
}

func (l *Leader) Validate(msg []byte, r *big.Int, s *big.Int) bool {
    log.Debug("Validate signature", "responseBitmap", l.responseBitmap, "commitBitmap", l.commitBitmap)
    if len(l.responseBitmap) < len(l.commitBitmap) {
        log.Debug("peer in responseBitmap and commitBitmap was not correct", "responseBitmap", len(l.responseBitmap), "commitBitmap", len(l.commitBitmap))
        return false
    }
    if float64(len(l.responseBitmap)) < math.Ceil(float64(len(l.members)*2.0/3.0)) {
        return false
    }
    sigmaPubKey := schnorr.CombinePubkeys(l.sigmaPubKey)
    return schnorr.Verify(sigmaPubKey, sha3.Hash256(msg), r, s)
}

func (l *Leader) getCommitNum() int {
    commitNum := 0
    for _, val := range l.commitBitmap {
        if val == 1 {
            commitNum = commitNum + 1
        }
    }
    return commitNum
}

func (l *Leader) getResponseNum() int {
    responseNum := 0
    for _, val := range l.responseBitmap {
        if val == 1 {
            responseNum = responseNum + 1
        }
    }
    return responseNum
}

func (l *Leader) setState(state int){
    l.stateLock.Lock()
    defer l.stateLock.Unlock()

    l.currentState = state
}

func (l *Leader) getState() int{
    l.stateLock.RLock()
    defer l.stateLock.RUnlock()

    return l.currentState
}