package service

import (
	"bytes"
	"github.com/drep-project/drep-chain/network/p2p/enode"
	"math/big"
	"sync"
	"time"

	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
	"github.com/drep-project/drep-chain/crypto/sha3"
	p2pService "github.com/drep-project/drep-chain/network/service"
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
	privakey *secp256k1.PrivateKey
	randomPrivakey *secp256k1.PrivateKey


	commitKey	*secp256k1.PublicKey
	p2pServer   p2pService.P2P

	commitBitmap      []byte
	sigmaPubKey       []*secp256k1.PublicKey
	sigmaCommitPubkey []*secp256k1.PublicKey
	msgHash           []byte

	sigmaS         *schnorr.Signature
	responseBitmap []byte
	syncLock       sync.Mutex

	waitTime time.Duration

	currentHeight       uint64
	minMember           int
	currentState        int
	stateLock           sync.RWMutex
	cancelWaitCommit    chan struct{}
	cancelWaitChallenge chan struct{}
}

func NewLeader(privkey *secp256k1.PrivateKey, p2pServer p2pService.P2P) *Leader {
	l := &Leader{}
	l.waitTime = 10 * time.Second
	l.pubkey = privkey.PubKey()
	l.privakey = privkey
	l.p2pServer = p2pServer

	l.Reset()
	return l
}

func (leader *Leader) UpdateStatus(producers []*consensusTypes.MemberInfo, minMember int, curHeight uint64) {
	leader.producers = producers
	leader.minMember = minMember
	leader.currentHeight = curHeight

	leader.liveMembers = []*consensusTypes.MemberInfo{}
	for _, producer := range producers {
		if producer.IsOnline && !producer.IsLeader {
			leader.liveMembers = append(leader.liveMembers, producer)
		}
	}
}

func (leader *Leader) Reset() {
	leader.sigmaPubKey = nil
	leader.sigmaCommitPubkey = nil
	leader.sigmaS = nil
	leader.randomPrivakey = nil

	length := len(leader.producers)
	leader.commitBitmap = make([]byte, length)
	leader.responseBitmap = make([]byte, length)

	leader.cancelWaitCommit = make(chan struct{}, 1)
	leader.cancelWaitChallenge = make(chan struct{}, 1)
}

func (leader *Leader) ProcessConsensus(msg []byte) (error, *secp256k1.Signature, []byte) {
	leader.setState(INIT)
	leader.setUp(msg)
	if !leader.waitForCommit() {
		//send reason and reset
		leader.fail("waitForCommit fail")
		return ErrWaitCommit, nil, nil
	}

	leader.challenge(msg)
	if !leader.waitForResponse() {
		//send reason and reset
		leader.fail("waitForResponse fail")
		return ErrWaitResponse, nil, nil
	}
	dlog.Debug("response complete")

	valid := leader.Validate(msg, leader.sigmaS.R, leader.sigmaS.S)
	dlog.Debug("vaidate result", "VALID", valid)
	if !valid {
		leader.fail("signature not valid")
		return ErrSignatureNotValid, nil, nil
	}
	return nil, &secp256k1.Signature{R: leader.sigmaS.R, S: leader.sigmaS.S}, leader.responseBitmap
}

func (leader *Leader) setUp(msg []byte) {
	setup := &consensusTypes.Setup{Msg: msg}
	setup.Height = leader.currentHeight
	leader.msgHash = sha3.Hash256(msg)
	var err error
	var nouncePk *secp256k1.PublicKey
	leader.randomPrivakey, nouncePk, err = schnorr.GenerateNoncePair(secp256k1.S256(), leader.msgHash, leader.privakey,nil, schnorr.Sha256VersionStringRFC6979)
	leader.sigmaPubKey = []*secp256k1.PublicKey{leader.pubkey}
	leader.sigmaCommitPubkey =  []*secp256k1.PublicKey{nouncePk}

	for i, v := range leader.producers {
		if v.Producer.Public.IsEqual(leader.pubkey) {
			leader.commitBitmap[i] = 1
		}
	}

	if err != nil {
		dlog.Error("generate private key error", "msg", err.Error())
		return
	}

	for _, member := range leader.liveMembers {
		if member.Peer != nil && !member.IsMe {
			dlog.Debug("leader sent setup message", "IP", member.Peer.IP(), "Height", setup.Height)
			leader.p2pServer.SendAsync(member.Peer.GetMsgRW(), consensusTypes.MsgTypeSetUp, setup)
		}
	}
}

func (leader *Leader) OnCommit(peer *consensusTypes.PeerInfo, commit *consensusTypes.Commitment) {
	leader.syncLock.Lock()
	defer leader.syncLock.Unlock()

	if leader.getState() != WAIT_COMMIT {
		return
	}
	if leader.currentHeight != commit.Height {
		return
	}

	leader.sigmaPubKey = append(leader.sigmaPubKey, commit.BpKey)
	leader.sigmaCommitPubkey = append(leader.sigmaCommitPubkey, commit.Q)

	leader.markCommit(peer)
	commitNum := leader.getCommitNum()
	if commitNum >= leader.minMember  {
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

func (leader *Leader) OnResponse(peer *consensusTypes.PeerInfo, response *consensusTypes.Response) {
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

	sigmaS, err := schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{leader.sigmaS, sig})
	if err != nil {
		dlog.Debug("schnorr combineSigs error", "reason", err)
		return
	} else {
		leader.sigmaS = sigmaS
		leader.markResponse(peer)
	}

	responseNum := leader.getResponseNum()
	if responseNum == len(leader.sigmaPubKey) {
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
	leader.selfSign(msg)
	for index, pk := range leader.sigmaPubKey {
		if index ==0 {
			continue
		}

		particateCommitPubkeys := []*secp256k1.PublicKey{}
		particateCommitPubkeys = append(particateCommitPubkeys, leader.sigmaCommitPubkey[0:index]...)
		particateCommitPubkeys = append(particateCommitPubkeys, leader.sigmaCommitPubkey[index+1:]...)

		commitPubkey := schnorr.CombinePubkeys(particateCommitPubkeys)
		challenge := &consensusTypes.Challenge{
			Height:      leader.currentHeight,
			SigmaQ:      commitPubkey,
			R:           leader.msgHash,
		}

		member := leader.getMemberByPk(pk)
		if member.IsOnline && !member.IsMe{
			dlog.Debug("leader sent challenge message", "IP", member.Peer.IP(), "Height", leader.currentHeight)
			leader.p2pServer.SendAsync(member.Peer.GetMsgRW(), consensusTypes.MsgTypeChallenge, challenge)
		}
	}
}

func (leader *Leader) selfSign(msg []byte) error {
	// pk1 | pk2 | pk3 | pk4
	commitPubkey := schnorr.CombinePubkeys(leader.sigmaCommitPubkey[1:])
	sig, err := schnorr.PartialSign(secp256k1.S256(), leader.msgHash, leader.privakey, leader.randomPrivakey, commitPubkey)
	if err != nil {
		return  err
	}
	leader.sigmaS = sig
	for i, v := range leader.producers {
		if v.Producer.Public.IsEqual(leader.pubkey) {
			leader.responseBitmap[i] = 1
		}
	}
	return nil
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
		if member.Peer != nil && !member.IsMe {
		leader.p2pServer.SendAsync(member.Peer.GetMsgRW(), consensusTypes.MsgTypeFail, failMsg)
		}
	}
}

func (leader *Leader) waitForResponse() bool {
	leader.setState(WAIT_RESPONSE)
	for {
		select {
		case <-time.After(leader.waitTime):
			responseNum := leader.getResponseNum()
			dlog.Debug("waitForResponse finish", "responseNum", responseNum, "liveMembers", len(leader.liveMembers))
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

func (leader *Leader) markResponse(peer *consensusTypes.PeerInfo) {
	index := leader.getMinerIndex(peer.GetID())
	if !leader.hasMarked(index, leader.responseBitmap) {
		return
	}
	leader.responseBitmap[index] = 1
}

func (leader *Leader) markCommit(peer *consensusTypes.PeerInfo) {
	index := leader.getMinerIndex(peer.GetID())
	if !leader.hasMarked(index, leader.commitBitmap) {
		return
	}
	leader.commitBitmap[index] = 1
}

func (leader *Leader) getMemberByIp(ip string) *consensusTypes.MemberInfo {
	for _, producer := range leader.producers {
		if producer.Peer != nil && producer.Peer.IP() == ip {
			return producer
		}
	}
	return nil
}

func (leader *Leader) getMemberByPk(pk *secp256k1.PublicKey) *consensusTypes.MemberInfo {
	for _, producer := range leader.producers {
		if producer.Peer != nil && producer.Producer.Public.IsEqual(pk)  {
			return producer
		}
	}
	return nil
}

func (leader *Leader) getMinerIndex(id *enode.ID) int {
	// TODO if it is itself
	for i, v := range leader.producers {
		if v.Peer != nil &&  bytes.Equal(v.Peer.GetID().Bytes(), id.Bytes()) {
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