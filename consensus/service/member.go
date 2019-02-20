package service

import (
    consensusTypes "github.com/drep-project/drep-chain/consensus/types"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "github.com/drep-project/drep-chain/log"
    p2pTypes "github.com/drep-project/drep-chain/network/types"
    p2pService "github.com/drep-project/drep-chain/network/service"
    "bytes"
    "errors"
    "math/big"
    "sync"
    "time"
)
const (
    TimeOoutEroor = "time out"
    LowHeightError = "leader's height  lower"
    HighHeightError = "leader's height  higher"
    StatusError = "error status"
    LeaderMistakeError = "setUp: mistake leader"
    ChallengeError =  "challenge error"
)

type Member struct {
    leader *p2pTypes.Peer
    members []*secp256k1.PublicKey
    prvKey *secp256k1.PrivateKey
    p2pServer *p2pService.P2pService

    msg []byte
    msgHash []byte

    randomPrivakey *secp256k1.PrivateKey
    r *big.Int

    waitTime time.Duration

    completed chan struct{}
    timeOutChanel chan struct{}
    errorChanel chan string
    cancelWaitSetUp  chan struct{}
    cancelWaitChallenge  chan struct{}
    currentState int
    currentHeight int64
    stateLock sync.RWMutex

    quitRound chan struct{}
}

func NewMember(prvKey *secp256k1.PrivateKey, quitRound chan struct{}, p2pServer *p2pService.P2pService) *Member {
    m := &Member{}

    m.prvKey = prvKey
    m.waitTime = 10 * time.Second

    m.p2pServer = p2pServer
    m.quitRound = quitRound

    m.Reset()
    return m
}

func (m *Member) UpdateStatus(participants []*p2pTypes.Peer , curMiner int,minMember int, curHeight int64){
    m.Reset()
    m.leader = participants[curMiner]
    m.members = []*secp256k1.PublicKey{}

    for _, participant := range participants {
        if participant == nil {
            m.members = append(m.members, m.prvKey.PubKey())
        }else {
            if !participant.PubKey.IsEqual(m.leader.PubKey) {
                m.members = append(m.members, participant.PubKey)
            }
        }
    }

    m.currentHeight = curHeight
}

func (m *Member) Reset(){
    m.msg  = nil
    m.msgHash = nil
    m.errorChanel = make(chan string,1)
    m.completed = make(chan struct{},1)
    m.cancelWaitSetUp = make(chan struct{},1)
    m.timeOutChanel = make(chan struct{},1)
    m.cancelWaitChallenge = make(chan struct{},1)
    m.setState(INIT)
}

func (m *Member) ProcessConsensus() ([]byte, error) {
    log.Debug("wait for leader's setup message", "IP",  m.leader.GetAddr())
    m.setState(WAIT_SETUP)
    go m.WaitSetUp()

    for {
        select {
        case msg := <- m.errorChanel:
            return nil, errors.New(msg)
        case <- m.timeOutChanel:
            m.setState(ERROR)
            return nil, errors.New(TimeOoutEroor)
        case <- m.completed:
            m.setState(COMPLETED)
            return m.msg, nil
        }
    }
}

func (m *Member) WaitSetUp(){
    select {
    case  <-time.After(m.waitTime):
        log.Debug("wait setup message timeout")
        m.setState(WAIT_SETUP_TIMEOUT)
        select {
        case m.timeOutChanel <- struct{}{}:
        default:
        }
        return
    case <- m.cancelWaitSetUp:
        return
    }
}

func (m *Member) OnSetUp(peer *p2pTypes.Peer, setUp *consensusTypes.Setup) {
    if m.currentHeight < setUp.Height {
        log.Debug("setup low height", "Receive Height", setUp.Height, "Current Height", m.currentHeight, "Status", m.getState())
        m.pushErrorMsg(HighHeightError)
        return
    }else if m.currentHeight > setUp.Height {
        log.Debug("setup high height", "Receive Height", setUp.Height, "Current Height", m.currentHeight, "Status", m.getState())
        m.pushErrorMsg(LowHeightError)
        return
    }

    if m.getState() != WAIT_SETUP{
        log.Debug("setup error status", "Receive Height", setUp.Height, "Current Height", m.currentHeight, "Status", m.getState())
        m.pushErrorMsg(StatusError)
        return
    }
    log.Debug("receive setup message")
    if m.leader.PubKey.IsEqual( peer.PubKey) {
        m.msg = setUp.Msg
        m.msgHash = sha3.Hash256(setUp.Msg)
        m.commit()
        log.Debug("sent commit message to leader")
        m.setState(WAIT_CHALLENGE)
        go m.WaitChallenge()
        select {
        case m.cancelWaitSetUp <- struct{}{}:
        default:
        }
    } else{
        //check fail not response and start new round
        m.pushErrorMsg(LeaderMistakeError)
    }
}

func (m *Member) WaitChallenge(){
    select {
    case  <-time.After(m.waitTime):
        m.setState(WAIT_CHALLENGE_TIMEOUT)
        select {
        case m.timeOutChanel <- struct{}{}:
        default:
        }
        return
    case <- m.cancelWaitChallenge:
        return
    }
}

func (m *Member) OnChallenge(peer *p2pTypes.Peer, challengeMsg *consensusTypes.Challenge) {
    if m.currentHeight < challengeMsg.Height {
        log.Debug("challenge high height", "Receive Height", challengeMsg.Height, "Current Height", m.currentHeight, "Status", m.getState())
        m.pushErrorMsg(HighHeightError)
        return
    }else if m.currentHeight > challengeMsg.Height {
        log.Debug("challenge high height", "Receive Height", challengeMsg.Height, "Current Height", m.currentHeight, "Status", m.getState())
        m.pushErrorMsg(LowHeightError)
        return
    }
    if m.getState() != WAIT_CHALLENGE{
        log.Debug("challenge error status", "Receive Height", challengeMsg.Height, "Current Height", m.currentHeight, "Status", m.getState())
        m.pushErrorMsg(StatusError)
        return
    }
    log.Debug("recieved challenge message")
    if m.leader.PubKey.IsEqual(peer.PubKey) {
        // log.Println("Member process challenge ")
        r := sha3.ConcatHash256(challengeMsg.SigmaPubKey.Serialize(), challengeMsg.SigmaQ.Serialize(), m.msgHash)
        //r0 := new(big.Int).SetBytes(challengeMsg.R)
        //rInt := new(big.Int).SetBytes(r)
        //curve := secp256k1.S256()
        //rInt.Mod(rInt, curve.Params().N)
        //m.r = rInt
        // if r0.Cmp(m.r) == 0{
        if bytes.Equal(r,challengeMsg.R) {
            m.response(challengeMsg)
            log.Debug("response has sent")
            m.setState(COMPLETED)
            select {
            case m.cancelWaitChallenge <- struct{}{}:
            default:
            }
            m.completed <- struct{}{}
            return
        }
    }
    m.pushErrorMsg(ChallengeError)
    //check fail not response and start new round
}

func (m *Member) OnFail(peer *p2pTypes.Peer, failMsg *consensusTypes.Fail){
    if m.currentHeight < failMsg.Height || m.getState() == COMPLETED || m.getState() == ERROR {
        return
    }
    log.Error("leader sent err message", "msg", failMsg.Reason)
    m.pushErrorMsg(failMsg.Reason)
}

func (m *Member) GetMembers() []*secp256k1.PublicKey{
    return m.members
}

func (m *Member) commit()  {
    var err error
    m.randomPrivakey, _, err = schnorr.GenerateNoncePair(secp256k1.S256(), m.msgHash, m.prvKey,nil, schnorr.Sha256VersionStringRFC6979)
    if err != nil {
        log.Error("generate private key error", "msg", err.Error())
        return
    }

    commitment := &consensusTypes.Commitment{Q: (*secp256k1.PublicKey)(&m.randomPrivakey.PublicKey)}
    commitment.Height = m.currentHeight
    m.p2pServer.SendAsync(m.leader, commitment)
}

func (m *Member) response(challengeMsg *consensusTypes.Challenge) {
    //  allPksSum1 := challengeMsg.SigmaQ.
    //  sig1, _ := schnorr.PartialSign(secp256k1.S256(), m.msgHash, m.prvKey, m.randomPrivakey, allPksSum1)

    //r := sha3.ConcatHash256(challengeMsg.SigmaQ.Serialize(), challengeMsg.SigmaPubKey.Serialize(), m.msg)
    sig, err := schnorr.PartialSign(secp256k1.S256(), m.msgHash, m.prvKey, m.randomPrivakey, challengeMsg.SigmaQ)
    if err != nil {
        log.Error("sign chanllenge error ", "msg", err.Error())
        return
    }
    response := &consensusTypes.Response{S: sig.Serialize()}
    response.Height = m.currentHeight
    m.p2pServer.SendAsync(m.leader, response)
}

func (m *Member) setState(state int){
    m.stateLock.Lock()
    defer m.stateLock.Unlock()

    m.currentState = state
}

func (m *Member) getState() int{
    m.stateLock.RLock()
    defer m.stateLock.RUnlock()

    return m.currentState
}

func (m *Member) pushErrorMsg(msg string) {
    m.setState(ERROR)
CANCEL:
    for{
        select {
        case m.errorChanel <- msg:
        case m.cancelWaitSetUp <- struct{}{}:
        case m.cancelWaitChallenge <- struct{}{}:
        default:
            break CANCEL
        }
    }
}