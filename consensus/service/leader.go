package service

import (
    "errors"
    "math"
    "math/big"
    "sync"
    "time"

    consensusTypes "github.com/drep-project/drep-chain/consensus/types"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "github.com/drep-project/drep-chain/log"
    p2pService "github.com/drep-project/drep-chain/network/service"
    p2pTypes "github.com/drep-project/drep-chain/network/types"
)

const (
    INIT = iota
    WAIT_SETUP
    WAIT_SETUP_TIMEOUT
    WAIT_COMMIT
    WAIT_COMMIT_COMPELED
    WAIT_COMMITT_IMEOUT
    WAIT_CHALLENGE
    WAIT_CHALLENGE_TIMEOUT
    WAIT_RESPONSE
    WAIT_RESPONSE_COMPELED
    WAIT_RESPONSE_TIMEOUT
    COMPLETED
    ERROR
)

type Leader struct {
    members    []*consensusTypes.Member
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

func (leader *Leader) UpdateStatus(members []*consensusTypes.Member, curMiner int, minMember int, curHeight int64){
    leader.members = make([]*consensusTypes.Member, len(members) - 1)
    leader.minMember = minMember //*2/3
    leader.currentHeight = curHeight

    last := -1
    for _, v := range members {
        if  v.Peer == nil {
            continue
        }
        last++
        leader.members[last] = v
    }
    leader.Reset()
}

func (leader *Leader) Reset(){
    leader.sigmaPubKey = nil
    leader.sigmaCommitPubkey = nil
    leader.sigmaS = nil
    length := len(leader.members)
    leader.commitBitmap = make([]byte, length)
    leader.responseBitmap = make([]byte, length)

    leader.cancelWaitCommit = make(chan struct{},1)
    leader.cancelWaitChallenge = make(chan struct{},1)
    leader.setState(INIT)
}


func (leader *Leader) ProcessConsensus(msg []byte) (error, *secp256k1.Signature, []byte) {
    leader.setState(INIT)
    //TODO  Waiting here is just to reduce the chance of message staggering. This waiting time should be minimized or even eliminated.
    time.AfterFunc(time.Second, func() {
        leader.setUp(msg)
    })
    if !leader.waitForCommit() {
        //send reason and reset
        leader.fail("waitForCommit fail")
        return errors.New("waitForCommit fail"), nil, nil
    }

    leader.challenge(msg)
    if !leader.waitForResponse() {
        //send reason and reset
        leader.fail("waitForResponse fail")
        return errors.New("waitForResponse fail"), nil, nil
    }
    log.Debug("response complete")

    if leader.sigmaS == nil {
        return errors.New("signature not valid"), nil,nil
    }
    valid := leader.Validate(msg, leader.sigmaS.R, leader.sigmaS.S)
    log.Debug("vaidate result","VALID", valid)
    if !valid && len(leader.members)>0 { //solo
        //return &util.ConnectionError{}, nil, nil
        return errors.New("signature not valid"), nil,nil
    }
    return nil, &secp256k1.Signature{R : leader.sigmaS.R, S : leader.sigmaS.S}, leader.responseBitmap
}

func (leader *Leader) setUp(msg []byte) {
    setup := &consensusTypes.Setup{ Msg: msg}
    setup.Height = leader.currentHeight
    for _, member := range leader.members {
        log.Debug("leader sent setup message", "IP", member.Peer.GetAddr(), "Height", setup.Height)
        leader.p2pServer.SendAsync(member.Peer, setup)
    }
}

func (leader *Leader) OnCommit(peer *p2pTypes.Peer, commit *consensusTypes.Commitment) {
    leader.syncLock.Lock()
    defer leader.syncLock.Unlock()

    if leader.getState() != WAIT_COMMIT {
        return
    }
    if leader.currentHeight != commit.Height {
        return
    }

    index := leader.getMinerIndex(peer.PubKey)
    if !hasMarked(index, leader.commitBitmap) {
        return
    }

    if leader.sigmaPubKey == nil {
        leader.sigmaPubKey = []*secp256k1.PublicKey{peer.PubKey }
    } else {
        leader.sigmaPubKey = append(leader.sigmaPubKey,  peer.PubKey)
    }

    if leader.sigmaCommitPubkey == nil {
        leader.sigmaCommitPubkey = []*secp256k1.PublicKey{commit.Q }
    }else{
        leader.sigmaCommitPubkey = append(leader.sigmaCommitPubkey, commit.Q)
    }

    leader.commitBitmap[index] = 1
    commitNum := leader.getCommitNum()
    if commitNum == len(leader.members){
        leader.setState(WAIT_COMMIT_COMPELED)
        log.Debug("OnCommit finish", "commitNum", commitNum, "members", len(leader.members))
        select {
        case leader.cancelWaitCommit <- struct{}{}:
        default:
        }
    }
}

func (leader *Leader) waitForCommit() bool {
    leader.setState(WAIT_COMMIT)
    for {
        select {
        case <-time.After(leader.waitTime):
            commitNum := leader.getCommitNum()
            log.Debug("waitForCommit  finish", "commitNum", commitNum, "members", len(leader.members))
            if commitNum >= leader.minMember  {
                return true
            }
            leader.setState(WAIT_COMMITT_IMEOUT)
            return false
        case <-leader.cancelWaitCommit:
            return true
        }
    }
}

func (leader *Leader) OnResponse(peer *p2pTypes.Peer, response *consensusTypes.Response) {
    leader.syncLock.Lock()
    defer leader.syncLock.Unlock()
    if leader.getState() != WAIT_RESPONSE {
        return
    }
    if leader.currentHeight != response.Height {
        return
    }

    sig, err := schnorr.ParseSignature(response.S)
    if err != nil {
        return
    }

    if leader.sigmaS == nil {
        leader.sigmaS = sig
        leader.markResponse(peer)
    }else{
        sigmaS, err := schnorr.CombineSigs(secp256k1.S256(),[]*schnorr.Signature{leader.sigmaS, sig })
        if err != nil {
            schnorr.CombineSigs(secp256k1.S256(),[]*schnorr.Signature{leader.sigmaS, sig })
            log.Debug("schnorr CombineSigs error", "reason", err)
            return
        }else {
            leader.sigmaS = sigmaS
            leader.markResponse(peer)
        }
    }

    responseNum := leader.getResponseNum()
    if responseNum == len(leader.members) && responseNum >= leader.minMember {
        leader.setState(WAIT_RESPONSE_COMPELED)
        log.Debug("OnResponse finish", "responseNum", responseNum, "members", len(leader.members))
        select {
        case  leader.cancelWaitChallenge <- struct{}{}:
        default:
            return
        }
    }
}

func (leader *Leader) markResponse(peer *p2pTypes.Peer) {
    index := leader.getMinerIndex(peer.PubKey)
    if !hasMarked(index, leader.responseBitmap) {
        return
    }
    leader.responseBitmap[index] = 1
}

func (leader *Leader) challenge(msg []byte) {
    for _, member := range leader.members {
        memIndex := 0
        sigmaPubKeys := []*secp256k1.PublicKey{}
        for index, pubkey := range  leader.sigmaPubKey {
            if !pubkey.IsEqual(member.Produce.Public) {
                sigmaPubKeys = append(sigmaPubKeys, pubkey)
            }else{
                memIndex = index
            }
        }
        sigmaPubKey := schnorr.CombinePubkeys(sigmaPubKeys)

        commitPubkeys := []*secp256k1.PublicKey{}
        for index, pubkey := range  leader.sigmaCommitPubkey {
            if memIndex != index {
                commitPubkeys = append(commitPubkeys, pubkey)
            }
        }
        commitPubkey := schnorr.CombinePubkeys(commitPubkeys)

        //leader.r = sha3.ConcatHash256(sigmaPubKey.Serialize(), commitPubkey.Serialize(), sha3.Hash256(msg))
        leader.r = sha3.ConcatHash256(sha3.Hash256(msg))
        challenge := &consensusTypes.Challenge{
            Height :      leader.currentHeight,
            SigmaPubKey : sigmaPubKey,
            SigmaQ :      commitPubkey,
            R:            leader.r,
        }
        log.Debug("leader sent challenge message", "IP", member.Peer.GetAddr(), "Height", leader.currentHeight)
        leader.p2pServer.SendAsync(member.Peer,challenge)
    }
}

func (leader *Leader) fail(msg string){
CANCEL:
    for{
        select {
        case leader.cancelWaitChallenge <- struct{}{}:
        case leader.cancelWaitCommit <- struct{}{}:
        default:
            break CANCEL
        }
    }
    failMsg := &consensusTypes.Fail{Reason: msg}
    failMsg.Height = leader.currentHeight
    for _, member := range leader.members {
        leader.p2pServer.SendAsync(member.Peer, failMsg)
    }
}

func (leader *Leader) waitForResponse() bool {
    leader.setState(WAIT_RESPONSE)
    for {
        select {
        case <-time.After(leader.waitTime):
            responseNum := leader.getResponseNum()
            log.Debug("waitForResponse finish", "responseNum", responseNum, "members", len(leader.members))
            if responseNum  >= leader.minMember{
                leader.setState(COMPLETED)
                return true
            }
            leader.setState(WAIT_RESPONSE_TIMEOUT)
            return false
        case <-leader.cancelWaitChallenge:
            return true
        }
    }
}

func (leader *Leader) getR(msg []byte) []byte {
    /*
    curve := secp256k1.S256()
    r := sha3.ConcatHash256(leader.sigmaCommitPubkey.Serialize(), leader.sigmaPubKey.Serialize(), msg)
    rInt := new(big.Int).SetBytes(r)
    rInt.Mod(rInt, curve.Params().N)
    return rInt.Bytes()
    */
    return []byte{}
}

func hasMarked(index int, bitmap []byte) bool {
    return index >=0 && index <= len(bitmap) && bitmap[index] != 1
}

func (leader *Leader) getMinerIndex(p *secp256k1.PublicKey) int {
    // TODO if it is itself
    for i, v := range leader.members {
        if v.Peer.PubKey.IsEqual(p) {
            return i
        }
    }
    return -1
}

func (leader *Leader) Validate(msg []byte, r *big.Int, s *big.Int) bool {
    log.Debug("Validate signature", "responseBitmap", leader.responseBitmap, "commitBitmap", leader.commitBitmap)
    if len(leader.responseBitmap) < len(leader.commitBitmap) {
        log.Debug("peer in responseBitmap and commitBitmap was not correct", "responseBitmap", len(leader.responseBitmap), "commitBitmap", len(leader.commitBitmap))
        return false
    }
    if float64(len(leader.responseBitmap)) < math.Ceil(float64(len(leader.members)*2.0/3.0)) {
        return false
    }

    sigmaPubKey := schnorr.CombinePubkeys(leader.getResponsePubkey())
    return schnorr.Verify(sigmaPubKey, sha3.Hash256(msg), r, s)
}

func (leader *Leader) getCommitNum() int {
    commitNum := 0
    for _, val := range leader.commitBitmap {
        if val == 1 {
            commitNum = commitNum + 1
        }
    }
    return commitNum
}

func (leader *Leader) getResponsePubkey() []*secp256k1.PublicKey {
    publicKeys := []*secp256k1.PublicKey{}
    for index, val := range leader.responseBitmap {
        if val == 1 {
            publicKeys = append(publicKeys, leader.members[index].Produce.Public)
        }
    }
    return publicKeys
}

func (leader *Leader) getResponseNum() int {
    responseNum := 0
    for _, val := range leader.responseBitmap {
        if val == 1 {
            responseNum = responseNum + 1
        }
    }
    return responseNum
}

func (leader *Leader) setState(state int){
    leader.stateLock.Lock()
    defer leader.stateLock.Unlock()

    leader.currentState = state
}

func (leader *Leader) getState() int{
    leader.stateLock.RLock()
    defer leader.stateLock.RUnlock()

    return leader.currentState
}