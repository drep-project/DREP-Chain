package service

import (
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
	"github.com/drep-project/drep-chain/crypto/sha3"
	p2pService "github.com/drep-project/drep-chain/network/service"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
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
	producers   []*consensusTypes.MemberInfo
	liveMembers []*consensusTypes.MemberInfo
	pubkey      *secp256k1.PublicKey
	p2pServer   p2pService.P2P

	commitBitmap      []byte
	sigmaPubKey       []*secp256k1.PublicKey
	sigmaCommitPubkey []*secp256k1.PublicKey
	r                 []byte

	sigmaS         *schnorr.Signature
	responseBitmap []byte
	syncLock       sync.Mutex

	waitTime time.Duration

	currentHeight       int64
	minMember           int
	currentState        int
	stateLock           sync.RWMutex
	cancelWaitCommit    chan struct{}
	cancelWaitChallenge chan struct{}
}

func NewLeader(pubKey *secp256k1.PublicKey, p2pServer p2pService.P2P) *Leader {
	l := &Leader{}
	l.waitTime = 10 * time.Second
	l.pubkey = pubKey
	l.p2pServer = p2pServer

	l.Reset()
	return l
}

func (leader *Leader) UpdateStatus(producers []*consensusTypes.MemberInfo, minMember int, curHeight int64) {
	leader.producers = producers
	leader.minMember = minMember
	leader.currentHeight = curHeight

	leader.liveMembers = []*consensusTypes.MemberInfo{}
	for _, producer := range producers {
		if producer.IsOnline && !producer.IsLeader && !producer.IsMe {
			leader.liveMembers = append(leader.liveMembers, producer)
		}
	}
}

func (leader *Leader) Reset() {
	leader.sigmaPubKey = nil
	leader.sigmaCommitPubkey = nil
	leader.sigmaS = nil
	length := len(leader.producers)
	leader.commitBitmap = make([]byte, length)
	leader.responseBitmap = make([]byte, length)

	leader.cancelWaitCommit = make(chan struct{}, 1)
	leader.cancelWaitChallenge = make(chan struct{}, 1)
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
	dlog.Debug("response complete")

	if leader.sigmaS == nil {
		return errors.New("signature not valid"), nil, nil
	}
	valid := leader.Validate(msg, leader.sigmaS.R, leader.sigmaS.S)
	dlog.Debug("vaidate result", "VALID", valid)
	if !valid {
		return errors.New("signature not valid"), nil, nil
	}
	return nil, &secp256k1.Signature{R: leader.sigmaS.R, S: leader.sigmaS.S}, leader.responseBitmap
}

func (leader *Leader) setUp(msg []byte) {
	setup := &consensusTypes.Setup{Msg: msg}
	setup.Height = leader.currentHeight
	for _, member := range leader.liveMembers {
		dlog.Debug("leader sent setup message", "IP", member.Peer.GetAddr(), "Height", setup.Height)
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

	member := leader.getMember(peer.Ip)
	if leader.sigmaPubKey == nil {
		leader.sigmaPubKey = []*secp256k1.PublicKey{member.Producer.Public}
	} else {
		leader.sigmaPubKey = append(leader.sigmaPubKey, member.Producer.Public)
	}

	if leader.sigmaCommitPubkey == nil {
		leader.sigmaCommitPubkey = []*secp256k1.PublicKey{commit.Q}
	} else {
		leader.sigmaCommitPubkey = append(leader.sigmaCommitPubkey, commit.Q)
	}

	leader.markCommit(peer)
	commitNum := leader.getCommitNum()
	if commitNum == len(leader.liveMembers) {
		leader.setState(WAIT_COMMIT_COMPELED)
		dlog.Debug("OnCommit finish", "commitNum", commitNum, "producers", len(leader.producers))
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
			dlog.Debug("waitForCommit  finish", "commitNum", commitNum, "producers", len(leader.producers))
			if commitNum >= leader.minMember {
				return true
			}
			leader.setState(WAIT_COMMIT_IMEOUT)
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
	} else {
		sigmaS, err := schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{leader.sigmaS, sig})
		if err != nil {
			schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{leader.sigmaS, sig})
			dlog.Debug("schnorr combineSigs error", "reason", err)
			return
		} else {
			leader.sigmaS = sigmaS
			leader.markResponse(peer)
		}
	}

	responseNum := leader.getResponseNum()
	if responseNum == len(leader.liveMembers) && responseNum >= leader.minMember {
		leader.setState(WAIT_RESPONSE_COMPELED)
		dlog.Debug("OnResponse finish", "responseNum", responseNum, "producers", len(leader.producers))
		select {
		case leader.cancelWaitChallenge <- struct{}{}:
		default:
			return
		}
	}
}

func (leader *Leader) challenge(msg []byte) {
	for _, member := range leader.liveMembers {
		memIndex := 0
		sigmaPubKeys := []*secp256k1.PublicKey{}
		for index, pubkey := range leader.sigmaPubKey {
			if !pubkey.IsEqual(member.Producer.Public) {
				sigmaPubKeys = append(sigmaPubKeys, pubkey)
			} else {
				memIndex = index
			}
		}
		sigmaPubKey := schnorr.CombinePubkeys(sigmaPubKeys)

		commitPubkeys := []*secp256k1.PublicKey{}
		for index, pubkey := range leader.sigmaCommitPubkey {
			if memIndex != index {
				commitPubkeys = append(commitPubkeys, pubkey)
			}
		}
		commitPubkey := schnorr.CombinePubkeys(commitPubkeys)

		//leader.r = sha3.ConcatHash256(sigmaPubKey.Serialize(), commitPubkey.Serialize(), sha3.Hash256(msg))
		leader.r = sha3.ConcatHash256(sha3.Hash256(msg))
		challenge := &consensusTypes.Challenge{
			Height:      leader.currentHeight,
			SigmaPubKey: sigmaPubKey,
			SigmaQ:      commitPubkey,
			R:           leader.r,
		}
		dlog.Debug("leader sent challenge message", "IP", member.Peer.GetAddr(), "Height", leader.currentHeight)
		leader.p2pServer.SendAsync(member.Peer, challenge)
	}
}

func (leader *Leader) fail(msg string) {
CANCEL:
	for {
		select {
		case leader.cancelWaitChallenge <- struct{}{}:
		case leader.cancelWaitCommit <- struct{}{}:
		default:
			break CANCEL
		}
	}
	failMsg := &consensusTypes.Fail{Reason: msg}
	failMsg.Height = leader.currentHeight
	for _, member := range leader.liveMembers {
		leader.p2pServer.SendAsync(member.Peer, failMsg)
	}
}

func (leader *Leader) waitForResponse() bool {
	leader.setState(WAIT_RESPONSE)
	for {
		select {
		case <-time.After(leader.waitTime):
			responseNum := leader.getResponseNum()
			dlog.Debug("waitForResponse finish", "responseNum", responseNum, "liveMembers", len(leader.liveMembers))
			if responseNum >= leader.minMember {
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

func (leader *Leader) Validate(msg []byte, r *big.Int, s *big.Int) bool {
	dlog.Debug("Validate signature", "responseBitmap", leader.responseBitmap, "commitBitmap", leader.commitBitmap)
	if len(leader.responseBitmap) < len(leader.commitBitmap) {
		dlog.Debug("peer in responseBitmap and commitBitmap was not correct", "responseBitmap", len(leader.responseBitmap), "commitBitmap", len(leader.commitBitmap))
		return false
	}
	if len(leader.responseBitmap) < leader.minMember {
		return false
	}

	sigmaPubKey := schnorr.CombinePubkeys(leader.getResponsePubkey())
	return schnorr.Verify(sigmaPubKey, sha3.Hash256(msg), r, s)
}

func (leader *Leader) hasMarked(index int, bitmap []byte) bool {
	return index >= 0 && index <= len(bitmap) && bitmap[index] != 1
}

func (leader *Leader) markResponse(peer *p2pTypes.Peer) {
	index := leader.getMinerIndex(peer.Ip)
	if !leader.hasMarked(index, leader.responseBitmap) {
		return
	}
	leader.responseBitmap[index] = 1
}

func (leader *Leader) markCommit(peer *p2pTypes.Peer) {
	index := leader.getMinerIndex(peer.Ip)
	if !leader.hasMarked(index, leader.commitBitmap) {
		return
	}
	leader.commitBitmap[index] = 1
}

func (leader *Leader) getMember(ip string) *consensusTypes.MemberInfo {
	for _, producer := range leader.producers {
		if producer.Peer.Ip == ip {
			return producer
		}
	}
	return nil
}

func (leader *Leader) getMinerIndex(ip string) int {
	// TODO if it is itself
	for i, v := range leader.producers {
		if v.Peer.Ip == ip {
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
			publicKeys = append(publicKeys, leader.producers[index].Producer.Public)
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
			publicKeys = append(publicKeys, leader.producers[index].Producer.Public)
		}
	}
	return publicKeys
}

func (leader *Leader) setState(state int) {
	leader.stateLock.Lock()
	defer leader.stateLock.Unlock()

	leader.currentState = state
}

func (leader *Leader) getState() int {
	leader.stateLock.RLock()
	defer leader.stateLock.RUnlock()

	return leader.currentState
}
