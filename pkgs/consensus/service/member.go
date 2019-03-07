package service

import (
    "sync"
    "time"
    "bytes"
    "errors"
    "math/big"

    "github.com/drep-project/dlog"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
    p2pService "github.com/drep-project/drep-chain/network/service"
    p2pTypes "github.com/drep-project/drep-chain/network/types"
    consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
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
    leader *consensusTypes.MemberInfo
    members []*secp256k1.PublicKey
    prvKey *secp256k1.PrivateKey
    p2pServer p2pService.P2P

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

    msgPool	chan *consensusTypes.RouteMsgWrap
    isConsensus bool   // time split 2, in consensus \ wait
}

func NewMember(prvKey *secp256k1.PrivateKey, p2pServer p2pService.P2P) *Member {
    member := &Member{}

    member.prvKey = prvKey
    member.waitTime = 10 * time.Second

    member.p2pServer = p2pServer
    member.msgPool = make(chan *consensusTypes.RouteMsgWrap, 1000)

    member.Reset()
    return member
}

func (member *Member) UpdateStatus(participants []*consensusTypes.MemberInfo , curMiner int,minMember int, curHeight int64){
    member.Reset()
    member.leader = participants[curMiner]
    member.members = []*secp256k1.PublicKey{}

    for _, participant := range participants {
        if participant.Peer == nil {
            member.members = append(member.members, member.prvKey.PubKey())
        }else {
            if !participant.Producer.Public.IsEqual(member.leader.Producer.Public) {
                member.members = append(member.members, participant.Producer.Public)
            }
        }
    }

    member.currentHeight = curHeight
}

func (member *Member) Reset(){
    member.msg  = nil
    member.msgHash = nil
    member.errorChanel = make(chan string,1)
    member.completed = make(chan struct{},1)
    member.cancelWaitSetUp = make(chan struct{},1)
    member.timeOutChanel = make(chan struct{},1)
    member.cancelWaitChallenge = make(chan struct{},1)
    member.setState(INIT)
}

func (member *Member) ProcessConsensus() ([]byte, error) {
    dlog.Debug("wait for leader's setup message", "IP",  member.leader.Peer.GetAddr())
    member.setState(WAIT_SETUP)
    member.isConsensus = true
    defer func() {
        member.isConsensus = false
    }()
    go member.WaitSetUp()

PREMSG:
    for {
        select {
        case msg := <- member.msgPool:
            member.OnSetUp(msg.Peer, msg.SetUpMsg)
        default:
            break PREMSG
        }
    }

    for {
        select {
        case msg := <- member.errorChanel:
            return nil, errors.New(msg)
        case <- member.timeOutChanel:
            member.setState(ERROR)
            return nil, errors.New(TimeOoutEroor)
        case <- member.completed:
            member.setState(COMPLETED)
            return member.msg, nil
        }
    }
}

func (member *Member) WaitSetUp(){
    select {
    case  <-time.After(member.waitTime):
        dlog.Debug("wait setup message timeout")
        member.setState(WAIT_SETUP_TIMEOUT)
        select {
        case member.timeOutChanel <- struct{}{}:
        default:
        }
        return
    case <- member.cancelWaitSetUp:
        return
    }
}

func (member *Member) OnSetUp(peer *p2pTypes.Peer, setUp *consensusTypes.Setup) {
    if !member.isConsensus {
        member.msgPool <- &consensusTypes.RouteMsgWrap{
            Peer: peer,
            SetUpMsg: setUp,
        }
        dlog.Debug("restore setup message")
        return
    }


    if member.currentHeight < setUp.Height {
        dlog.Debug("setup low height", "Receive Height", setUp.Height, "Current Height", member.currentHeight, "Status", member.getState())
        member.pushErrorMsg(HighHeightError)
        return
    }else if member.currentHeight > setUp.Height {
        dlog.Debug("setup high height", "Receive Height", setUp.Height, "Current Height", member.currentHeight, "Status", member.getState())
        member.pushErrorMsg(LowHeightError)
        return
    }

    if member.getState() != WAIT_SETUP{
        dlog.Debug("setup error status", "Receive Height", setUp.Height, "Current Height", member.currentHeight, "Status", member.getState())
        member.pushErrorMsg(StatusError)
        return
    }
    dlog.Debug("receive setup message")
    if member.leader.Peer.Ip == peer.Ip {
        member.msg = setUp.Msg
        member.msgHash = sha3.Hash256(setUp.Msg)
        member.commit()
        dlog.Debug("sent commit message to leader")
        member.setState(WAIT_CHALLENGE)
        go member.WaitChallenge()
        select {
        case member.cancelWaitSetUp <- struct{}{}:
        default:
        }
    } else{
        //check fail not response and start new round
        member.pushErrorMsg(LeaderMistakeError)
    }
}

func (member *Member) WaitChallenge(){
    select {
    case  <-time.After(member.waitTime):
        member.setState(WAIT_CHALLENGE_TIMEOUT)
        select {
        case member.timeOutChanel <- struct{}{}:
        default:
        }
        return
    case <- member.cancelWaitChallenge:
        return
    }
}

func (member *Member) OnChallenge(peer *p2pTypes.Peer, challengeMsg *consensusTypes.Challenge) {
    if member.currentHeight < challengeMsg.Height {
        dlog.Debug("challenge high height", "Receive Height", challengeMsg.Height, "Current Height", member.currentHeight, "Status", member.getState())
        member.pushErrorMsg(HighHeightError)
        return
    }else if member.currentHeight > challengeMsg.Height {
        dlog.Debug("challenge high height", "Receive Height", challengeMsg.Height, "Current Height", member.currentHeight, "Status", member.getState())
        member.pushErrorMsg(LowHeightError)
        return
    }
    if member.getState() != WAIT_CHALLENGE{
        dlog.Debug("challenge error status", "Receive Height", challengeMsg.Height, "Current Height", member.currentHeight, "Status", member.getState())
        member.pushErrorMsg(StatusError)
        return
    }
    dlog.Debug("recieved challenge message")
    if member.leader.Peer.Ip == peer.Ip {
        // r := sha3.ConcatHash256(challengeMsg.SigmaPubKey.Serialize(), challengeMsg.SigmaQ.Serialize(), member.msgHash)
        r := sha3.ConcatHash256(member.msgHash)
        // dlog.Println("Member process challenge ")
        //r0 := new(big.Int).SetBytes(challengeMsg.R)
        //rInt := new(big.Int).SetBytes(r)
        //curve := secp256k1.S256()
        //rInt.Mod(rInt, curve.Params().N)
        //member.r = rInt
        // if r0.Cmp(member.r) == 0{
        if bytes.Equal(r,challengeMsg.R) {
            member.response(challengeMsg)
            dlog.Debug("response has sent")
            member.setState(COMPLETED)
            select {
            case member.cancelWaitChallenge <- struct{}{}:
            default:
            }
            member.completed <- struct{}{}
            return
        }
    }
    member.pushErrorMsg(ChallengeError)
    //check fail not response and start new round
}

func (member *Member) OnFail(peer *p2pTypes.Peer, failMsg *consensusTypes.Fail){
    if member.currentHeight < failMsg.Height || member.getState() == COMPLETED || member.getState() == ERROR {
        return
    }
    dlog.Error("member receive leader's err message", "msg", failMsg.Reason)
    member.pushErrorMsg(failMsg.Reason)
}

func (member *Member) GetMembers() []*secp256k1.PublicKey{
    return member.members
}

func (member *Member) commit()  {
    //TODO validate block from leader
    var err error
    member.randomPrivakey, _, err = schnorr.GenerateNoncePair(secp256k1.S256(), member.msgHash, member.prvKey,nil, schnorr.Sha256VersionStringRFC6979)
    if err != nil {
        dlog.Error("generate private key error", "msg", err.Error())
        return
    }

    commitment := &consensusTypes.Commitment{Q: (*secp256k1.PublicKey)(&member.randomPrivakey.PublicKey)}
    commitment.Height = member.currentHeight
    member.p2pServer.SendAsync(member.leader.Peer, commitment)
}

func (member *Member) response(challengeMsg *consensusTypes.Challenge) {
    //  allPksSum1 := challengeMsg.SigmaQ.
    //  sig1, _ := schnorr.PartialSign(secp256k1.S256(), member.msgHash, member.prvKey, member.randomPrivakey, allPksSum1)

    //r := sha3.ConcatHash256(challengeMsg.SigmaQ.Serialize(), challengeMsg.SigmaPubKey.Serialize(), member.msg)
    sig, err := schnorr.PartialSign(secp256k1.S256(), member.msgHash, member.prvKey, member.randomPrivakey, challengeMsg.SigmaQ)
    if err != nil {
        dlog.Error("sign chanllenge error ", "msg", err.Error())
        return
    }
    response := &consensusTypes.Response{S: sig.Serialize()}
    response.Height = member.currentHeight
    member.p2pServer.SendAsync(member.leader.Peer, response)
}

func (member *Member) setState(state int){
    member.stateLock.Lock()
    defer member.stateLock.Unlock()

    member.currentState = state
}

func (member *Member) getState() int{
    member.stateLock.RLock()
    defer member.stateLock.RUnlock()

    return member.currentState
}

func (member *Member) pushErrorMsg(msg string) {
    member.setState(ERROR)
CANCEL:
    for{
        select {
        case member.errorChanel <- msg:
        case member.cancelWaitSetUp <- struct{}{}:
        case member.cancelWaitChallenge <- struct{}{}:
        default:
            break CANCEL
        }
    }
}
