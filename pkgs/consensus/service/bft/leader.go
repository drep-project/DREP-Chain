package bft

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1/schnorr"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	consensusTypes "github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	"github.com/drep-project/binary"
	"math/big"
	"sync"
	"time"
)

const (
	INIT = iota
	WAIT_SETUP
	WAIT_SETUP_TIMEOUT
	WAIT_COMMIT
	WAIT_COMMIT_COMPELED
	WAIT_COMMIT_IMEOUT
	WAIT_CHALLENGE
	WAIT_CHALLENGE_TIMEOUT
	WAIT_RESPONSE
	WAIT_RESPONSE_COMPELED
	WAIT_RESPONSE_TIMEOUT
	COMPLETED
	ERROR
)

type Leader struct {
	producers   []*MemberInfo
	liveMembers []*MemberInfo

	pubkey         *secp256k1.PublicKey
	privakey       *secp256k1.PrivateKey
	randomPrivakey *secp256k1.PrivateKey

	commitKey *secp256k1.PublicKey
	sender    Sender

	commitBitmap      []byte
	sigmaPubKey       []*secp256k1.PublicKey
	sigmaCommitPubkey []*secp256k1.PublicKey
	msgHash           []byte

	sigmaS         *schnorr.Signature
	responseBitmap []byte
	syncLock       sync.Mutex

	msgPool    chan *MsgWrap
	cancelPool chan struct{}
	waitTime   time.Duration

	currentHeight       uint64
	minMember           int
	currentState        int
	stateLock           sync.RWMutex
	cancelWaitCommit    chan struct{}
	cancelWaitChallenge chan struct{}
}

func NewLeader(privkey *secp256k1.PrivateKey, p2pServer Sender, waitTime time.Duration, producers []*MemberInfo, minMember int, curHeight uint64, msgPool chan *MsgWrap) *Leader {
	l := &Leader{}
	l.pubkey = privkey.PubKey()
	l.privakey = privkey
	l.waitTime = waitTime
	l.sender = p2pServer
	l.msgPool = msgPool
	l.producers = producers
	l.minMember = minMember
	l.currentHeight = curHeight

	l.liveMembers = []*MemberInfo{}
	for _, producer := range producers {
		if producer.IsOnline && !producer.IsLeader {
			l.liveMembers = append(l.liveMembers, producer)
		}
	}
	l.cancelPool = make(chan struct{})
	l.Reset()
	return l
}

func (leader *Leader) Reset() {
	leader.sigmaPubKey = nil
	leader.sigmaCommitPubkey = nil
	leader.sigmaS = nil
	leader.randomPrivakey = nil

	length := len(leader.producers)
	leader.commitBitmap = make([]byte, length)
	leader.responseBitmap = make([]byte, length)

	leader.cancelWaitCommit = make(chan struct{})
	leader.cancelWaitChallenge = make(chan struct{})
}

func (leader *Leader) Close() {
	close(leader.cancelWaitCommit)
	close(leader.cancelWaitChallenge)
}

/*

leader                   member
setup      ----->
	       <-----      commit
challenge  ----->
           <-----      response
*/
func (leader *Leader) ProcessConsensus(msg IConsenMsg, round int) (error, *secp256k1.Signature, []byte) {
	defer func() {
		leader.cancelPool <- struct{}{}
	}()

	leader.setState(INIT)
	go leader.processP2pMessage(round)
	leader.setUp(msg, round)
	if !leader.waitForCommit() {
		//send reason and reset
		leader.fail(ErrWaitCommit.Error(), round)
		return ErrWaitCommit, nil, nil
	}
	leader.challenge(msg, round)

	if !leader.waitForResponse() {
		//send reason and reset
		leader.fail("waitForResponse fail", round)
		return ErrWaitResponse, nil, nil
	}
	log.Debug("response complete")
	valid := leader.Validate(msg, leader.sigmaS.R, leader.sigmaS.S)
	log.WithField("VALID", valid).Debug("vaidate result")
	if !valid {
		leader.fail("signature not valid", round)
		return ErrSignatureNotValid, nil, nil
	}
	return nil, &secp256k1.Signature{R: leader.sigmaS.R, S: leader.sigmaS.S}, leader.responseBitmap
}

func (leader *Leader) processP2pMessage(round int) {
	for {
		select {
		case msg := <-leader.msgPool:
			switch msg.Code {
			case MsgTypeCommitment:
				var req Commitment
				if err := binary.Unmarshal(msg.Msg, &req); err != nil {
					log.Debugf("commit msg:%v err:%v", msg, err)
					continue
				}

				if req.Round != round {
					log.WithField("come round", req.Round).WithField("local round", round).Info("leader process msg req err")
					continue
				}

				leader.OnCommit(msg.Peer, &req)
			case MsgTypeResponse:
				var res Response
				if err := binary.Unmarshal(msg.Msg, &res); err != nil {
					log.Debugf("response msg:%v err:%v", msg, err)
					continue
				}
				if res.Round != round {
					log.WithField("come round", res.Round).WithField("local round", round).Info("leader process msg res err")
					continue
				}

				leader.OnResponse(msg.Peer, &res)
			}
		case <-leader.cancelPool:
			return
		}
	}
}

func (leader *Leader) setUp(msg IConsenMsg, round int) {
	setup := &Setup{Msg: msg.AsMessage()}
	setup.Height = leader.currentHeight
	setup.Magic = SetupMagic
	setup.Round = round
	leader.msgHash = sha3.Keccak256(msg.AsSignMessage())
	var err error
	var nouncePk *secp256k1.PublicKey
	leader.randomPrivakey, nouncePk, err = schnorr.GenerateNoncePair(secp256k1.S256(), leader.msgHash, leader.privakey, nil, schnorr.Sha256VersionStringRFC6979)
	leader.sigmaPubKey = []*secp256k1.PublicKey{leader.pubkey}
	leader.sigmaCommitPubkey = []*secp256k1.PublicKey{nouncePk}

	for i, v := range leader.producers {
		if v.Producer.Pubkey.IsEqual(leader.pubkey) {
			leader.commitBitmap[i] = 1
		}
	}

	if err != nil {
		log.WithField("msg", err).Error("generate private key error")
		return
	}

	for _, member := range leader.liveMembers {
		if member.Peer != nil && !member.IsMe {
			log.WithField("Node", member.Peer.IP()).WithField("Height", setup.Height).WithField("size", len(setup.Msg)).Trace("leader sent setup message")
			leader.sender.SendAsync(member.Peer.GetMsgRW(), MsgTypeSetUp, setup)
		}
	}
}

func (leader *Leader) OnCommit(peer consensusTypes.IPeerInfo, commit *Commitment) {
	leader.syncLock.Lock()
	defer leader.syncLock.Unlock()

	if leader.getState() != WAIT_COMMIT {
		log.WithField("current status", leader.getState()).WithField("receive message", commit).Debug("wrong commit message state")
		return
	}
	if leader.currentHeight != commit.Height {
		log.WithField("current height", leader.currentHeight).WithField("receive message", commit).Debug("wrong commit message state")
		return
	}

	leader.sigmaPubKey = append(leader.sigmaPubKey, commit.BpKey)
	leader.sigmaCommitPubkey = append(leader.sigmaCommitPubkey, commit.Q)

	leader.markCommit(peer)
	commitNum := leader.getCommitNum()
	if commitNum >= leader.minMember {
		leader.setState(WAIT_COMMIT_COMPELED)
		log.WithField("commitNum", commitNum).WithField("producers", len(leader.producers)).Debug("OnCommit finish")
		select {
		case leader.cancelWaitCommit <- struct{}{}:
		default:
		}
	}
}

func (leader *Leader) waitForCommit() bool {
	leader.setState(WAIT_COMMIT)
	//fmt.Println(leader.waitTime.String())
	t := time.Now()
	tm := time.NewTimer(leader.waitTime)
	select {
	case <-tm.C:
		commitNum := leader.getCommitNum()
		log.WithField("commitNum", commitNum).WithField("producers", len(leader.producers)).Debug("waitForCommit  finish")
		log.WithField("start", t).WithField("now", time.Now()).Info("wait for commit timeout")
		if commitNum >= leader.minMember {
			return true
		}
		leader.setState(WAIT_COMMIT_IMEOUT)
		return false
	case <-leader.cancelWaitCommit:
		return true
	}
}

func (leader *Leader) OnResponse(peer consensusTypes.IPeerInfo, response *Response) {
	leader.syncLock.Lock()
	defer leader.syncLock.Unlock()
	if leader.getState() != WAIT_RESPONSE {
		log.WithField("current status", leader.getState()).WithField("receive message", response).Debug("wrong response message state")
		return
	}
	if leader.currentHeight != response.Height {
		log.WithField("current height", leader.currentHeight).WithField("receive message", response).Debug("wrong response message height")
		return
	}

	sig, err := schnorr.ParseSignature(response.S)
	if err != nil {
		return
	}

	sigmaS, err := schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{leader.sigmaS, sig})
	if err != nil {
		log.WithField("reason", err).Debug("schnorr combineSigs error")
		return
	} else {
		leader.sigmaS = sigmaS
		leader.markResponse(peer)
	}

	responseNum := leader.getResponseNum()
	if responseNum == len(leader.sigmaPubKey) {
		leader.setState(WAIT_RESPONSE_COMPELED)
		log.WithField("responseNum", responseNum).WithField("producers", len(leader.producers)).Debug("OnResponse finish")
		select {
		case leader.cancelWaitChallenge <- struct{}{}:
		default:
			return
		}
	}
}

func (leader *Leader) challenge(msg IConsenMsg, Round int) {
	leader.selfSign(msg)
	for index, pk := range leader.sigmaPubKey {
		if index == 0 {
			continue
		}

		particateCommitPubkeys := []*secp256k1.PublicKey{}
		particateCommitPubkeys = append(particateCommitPubkeys, leader.sigmaCommitPubkey[0:index]...)
		particateCommitPubkeys = append(particateCommitPubkeys, leader.sigmaCommitPubkey[index+1:]...)

		commitPubkey := schnorr.CombinePubkeys(particateCommitPubkeys)
		challenge := &Challenge{
			Height: leader.currentHeight,
			Magic:  ChallegeMagic,
			Round:  Round,
			SigmaQ: commitPubkey,
			R:      leader.msgHash,
		}

		member := leader.getMemberByPk(pk)
		if member != nil && member.IsOnline && !member.IsMe {
			log.WithField("Node", member.Peer.IP()).WithField("Height", leader.currentHeight).Debug("leader sent challenge message")
			leader.sender.SendAsync(member.Peer.GetMsgRW(), MsgTypeChallenge, challenge)
		}
	}
}

func (leader *Leader) selfSign(msg IConsenMsg) error {
	// pk1 | pk2 | pk3 | pk4
	commitPubkey := schnorr.CombinePubkeys(leader.sigmaCommitPubkey[1:])
	sig, err := schnorr.PartialSign(secp256k1.S256(), leader.msgHash, leader.privakey, leader.randomPrivakey, commitPubkey)
	if err != nil {
		return err
	}
	leader.sigmaS = sig
	for i, v := range leader.producers {
		if v.Producer.Pubkey.IsEqual(leader.pubkey) {
			leader.responseBitmap[i] = 1
		}
	}
	return nil
}

func (leader *Leader) fail(msg string, round int) {
CANCEL:
	for {
		select {
		case leader.cancelWaitChallenge <- struct{}{}:
		case leader.cancelWaitCommit <- struct{}{}:
		default:
			break CANCEL
		}
	}
	failMsg := &Fail{Reason: msg, Magic: FailMagic, Round: round}
	failMsg.Height = leader.currentHeight

	for _, member := range leader.liveMembers {
		if member.Peer != nil && !member.IsMe {
			leader.sender.SendAsync(member.Peer.GetMsgRW(), MsgTypeFail, failMsg)
		}
	}
}

func (leader *Leader) waitForResponse() bool {
	leader.setState(WAIT_RESPONSE)
	tm := time.NewTimer(leader.waitTime)
	select {
	case <-tm.C:
		responseNum := leader.getResponseNum()
		log.WithField("responseNum", responseNum).WithField("liveMembers", len(leader.liveMembers)).Debug("waitForResponse finish")
		if responseNum == len(leader.sigmaPubKey) {
			leader.setState(COMPLETED)
			return true
		}
		leader.setState(WAIT_RESPONSE_TIMEOUT)
		return false
	case <-leader.cancelWaitChallenge:
		return true
	}

}

func (leader *Leader) Validate(msg IConsenMsg, r *big.Int, s *big.Int) bool {
	log.WithField("responseBitmap", leader.responseBitmap).WithField("commitBitmap", leader.commitBitmap).Debug("Validate signature")
	if len(leader.responseBitmap) < len(leader.commitBitmap) {
		log.WithField("responseBitmap", len(leader.responseBitmap)).WithField("commitBitmap", len(leader.commitBitmap)).Debug("peer in responseBitmap and commitBitmap was not correct")
		return false
	}
	if len(leader.responseBitmap) < leader.minMember {
		return false
	}
	sigmaPubKey := schnorr.CombinePubkeys(leader.getResponsePubkey())
	return schnorr.Verify(sigmaPubKey, sha3.Keccak256(msg.AsSignMessage()), r, s)
}

func (leader *Leader) hasMarked(index int, bitmap []byte) bool {
	return index >= 0 && index <= len(bitmap) && bitmap[index] != 1
}

func (leader *Leader) markResponse(peer consensusTypes.IPeerInfo) {
	index := leader.getMinerIndex(peer)
	if !leader.hasMarked(index, leader.responseBitmap) {
		return
	}
	leader.responseBitmap[index] = 1
}

func (leader *Leader) markCommit(peer consensusTypes.IPeerInfo) {
	index := leader.getMinerIndex(peer)
	if !leader.hasMarked(index, leader.commitBitmap) {
		return
	}
	leader.commitBitmap[index] = 1
}

func (leader *Leader) getMemberByPk(pk *secp256k1.PublicKey) *MemberInfo {
	for _, producer := range leader.producers {
		if producer.Peer != nil && producer.Producer.Pubkey.IsEqual(pk) {
			return producer
		}
	}
	return nil
}

func (leader *Leader) getMinerIndex(peer consensusTypes.IPeerInfo) int {
	for i, v := range leader.producers {
		if v.Peer != nil && v.Peer.Equal(peer) {
			return i
		}
	}
	return -1
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

func (leader *Leader) getCommitPubkey() []*secp256k1.PublicKey {
	publicKeys := []*secp256k1.PublicKey{}
	for index, val := range leader.commitBitmap {
		if val == 1 {
			publicKeys = append(publicKeys, leader.producers[index].Producer.Pubkey)
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

func (leader *Leader) getResponsePubkey() []*secp256k1.PublicKey {
	publicKeys := []*secp256k1.PublicKey{}
	for index, val := range leader.responseBitmap {
		if val == 1 {
			publicKeys = append(publicKeys, leader.producers[index].Producer.Pubkey)
		}
	}
	return publicKeys
}

func (leader *Leader) setState(state int) {
	leader.stateLock.Lock()
	defer leader.stateLock.Unlock()

	if state == WAIT_COMMIT_IMEOUT {
		fmt.Print("")
	}
	leader.currentState = state
}

func (leader *Leader) getState() int {
	leader.stateLock.RLock()
	defer leader.stateLock.RUnlock()

	return leader.currentState
}
