package bft

import (
	"fmt"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/network/p2p"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.TraceLevel)
}


type testSendor struct {
	onlinePeers map[string]consensusTypes.IPeerInfo
	clients     []*testBFT
	localIp     string
}

func (testSendor *testSendor) SendAsync(w p2p.MsgWriter, msgType uint64, msg interface{}) chan error {
	go func() {
		ll := ((interface{})(w)).(*writeIo)
		bytes, _ := binary.Marshal(msg)
		ll.peer.client.ReceiveMsg(testSendor.onlinePeers[testSendor.localIp], msgType, bytes)
	}()
	return nil
}

type writeIo struct {
	peer *testPeer
}

func (writeIo *writeIo) ReadMsg() (p2p.Msg, error) {
	return p2p.Msg{}, nil
}

func (writeIo *writeIo) WriteMsg(msg p2p.Msg) error {
	return nil
}

type testPeer struct {
	consensusTypes.Producer
	client *testBFT
}

func (testPeer *testPeer) GetMsgRW() p2p.MsgReadWriter {
	return &writeIo{testPeer}
}

func (testPeer *testPeer) String() string {
	return testPeer.Producer.IP
}

func (testPeer *testPeer) IP() string {
	return testPeer.Producer.IP
}

func (testPeer *testPeer) Equal(ipeer consensusTypes.IPeerInfo) bool {
	return testPeer.Producer.IP == ipeer.IP()
}

type dummyConsensusMsg struct {
}

func (dummyConsensusMsg *dummyConsensusMsg) AsSignMessage() []byte {
	return []byte{}
}

func (dummyConsensusMsg *dummyConsensusMsg) AsMessage() []byte {
	return []byte{}
}


type testBFT struct {
	PrivKey      *secp256k1.PrivateKey
	curMiner     int
	minMiners    int
	curentHeight uint64
	onLinePeer   map[string]consensusTypes.IPeerInfo
	WaitTime     time.Duration
	sender       Sender
	ip string
	memberMsgPool chan *MsgWrap
	leaderMsgPool chan *MsgWrap

	leader    *Leader
	member    *Member
	Producers consensusTypes.ProducerSet
}

func newTestBFT(
	privKey *secp256k1.PrivateKey,
	producer consensusTypes.ProducerSet,
	sender Sender,
	ip string,
	peersInfo map[string]consensusTypes.IPeerInfo) *testBFT {
	return &testBFT{
		PrivKey:       privKey,
		minMiners:     int(math.Ceil(float64(len(producer)) * 2 / 3)),
		Producers:     producer,
		sender:        sender,
		ip:ip,
		onLinePeer:    peersInfo,
		WaitTime:      waitTime,
		memberMsgPool: make(chan *MsgWrap, 1000),
		leaderMsgPool: make(chan *MsgWrap, 1000),
	}
}

func (testbft *testBFT) Run() *bftResult {
	miners := testbft.collectMemberStatus()
	if len(miners) > 1 {
		isM, isL := testbft.moveToNextMiner(miners)
		if isL {
			return testbft.runAsLeader(miners)
		} else if isM {
			return testbft.runAsMember(miners)
		} else {
			return &bftResult{err:ErrBFTNotReady}
		}
	} else {
		return &bftResult{err:ErrBFTNotReady}
	}
}

func (testbft *testBFT) moveToNextMiner(produceInfos []*MemberInfo) (bool, bool) {
	liveMembers := []*MemberInfo{}

	for _, produce := range produceInfos {
		if produce.IsOnline {
			liveMembers = append(liveMembers, produce)
		}
	}
	liveMinerIndex := int(testbft.curentHeight % uint64(len(liveMembers)))
	curMiner := liveMembers[liveMinerIndex]

	for index, produce := range produceInfos {
		if produce.IsOnline {
			if produce.Producer.Pubkey.IsEqual(curMiner.Producer.Pubkey) {
				produce.IsLeader = true
				testbft.curMiner = index
			} else {
				produce.IsLeader = false
			}
		}
	}

	if curMiner.IsMe {
		return false, true
	} else {
		return true, false
	}
}

func (testbft *testBFT) collectMemberStatus() []*MemberInfo {
	produceInfos := make([]*MemberInfo, 0, len(testbft.Producers))
	for _, produce := range testbft.Producers {
		var (
			IsOnline, ok bool
			pi           consensusTypes.IPeerInfo
		)

		isMe := testbft.PrivKey.PubKey().IsEqual(produce.Pubkey)
		if isMe {
			IsOnline = true
		} else {
			//todo  peer获取到的IP地址和配置的ip地址是否相等（nat后是否相等,从tcp原理来看是相等的）
			if pi, ok = testbft.onLinePeer[produce.IP]; ok {
				IsOnline = true
			}
		}

		produceInfos = append(produceInfos, &MemberInfo{
			Producer: &consensusTypes.Producer{Pubkey: produce.Pubkey, IP: produce.IP},
			Peer:     pi,
			IsMe:     isMe,
			IsOnline: IsOnline,
		})
	}
	return produceInfos
}

type bftResult struct {
	bitmap   []byte
	msg      IConsenMsg
	multiSig *secp256k1.Signature
	err      error
}

func (testbft *testBFT) runAsLeader(miners []*MemberInfo) *bftResult {
	testbft.leader = NewLeader(testbft.PrivKey, testbft.sender, testbft.WaitTime, miners, testbft.minMiners, testbft.curentHeight, testbft.leaderMsgPool)
	err, sig, bitmap := testbft.leader.ProcessConsensus(&dummyConsensusMsg{})
	return &bftResult{bitmap, &dummyConsensusMsg{}, sig, err}
}

func (testbft *testBFT) runAsMember(miners []*MemberInfo) *bftResult {
	testbft.member = NewMember(testbft.PrivKey, testbft.sender, testbft.WaitTime, miners, testbft.minMiners, testbft.curentHeight, testbft.memberMsgPool)
	testbft.member.convertor = func(msg []byte) (IConsenMsg, error) {
		return &dummyConsensusMsg{}, nil
	}
	testbft.member.validator = func(msg IConsenMsg) bool {
		return true
	}
	msg, err := testbft.member.ProcessConsensus()
	return &bftResult{
		err: err,
		msg: msg,
	}
}

func (testbft *testBFT) ReceiveMsg(peer consensusTypes.IPeerInfo, code uint64, msg []byte) error {
	switch code {
	case MsgTypeSetUp:
		fallthrough
	case MsgTypeChallenge:
		fallthrough
	case MsgTypeFail:
		testbft.memberMsgPool <- &MsgWrap{peer, code, msg}
	case MsgTypeCommitment:
		fallthrough
	case MsgTypeResponse:
		testbft.leaderMsgPool <- &MsgWrap{peer, code, msg}
	default:
		return fmt.Errorf("consensus unkonw msg type:%d", code)
	}
	return nil
}

func (testbft *testBFT) ChangeTime(interval time.Duration) {
	testbft.WaitTime = interval
}

func TestBFT(t *testing.T) {
	keystore := make([]*secp256k1.PrivateKey, 4)
	produces := make([]consensusTypes.Producer, 4)
	onlinePeers := make(map[string]consensusTypes.IPeerInfo)
	bftClients := make([]*testBFT, 4)
	for i := 0; i < 4; i++ {
		priv, err := secp256k1.GeneratePrivateKey(nil)
		if err != nil {
			i--
			continue
		}
		p := consensusTypes.Producer{priv.PubKey(), strconv.Itoa(i)}
		produces[i] = p
		keystore[i] = priv

		sendor := &testSendor{onlinePeers, bftClients, p.IP}
		bftClient := newTestBFT(keystore[i], produces, sendor, p.IP, onlinePeers)
		bftClients[i] = bftClient
		onlinePeers[strconv.Itoa(i)] = &testPeer{p, bftClient}
	}

	group := &sync.WaitGroup{}
	for _, client := range bftClients {
		group.Add(1)
		go func(clientt *testBFT) {
			defer group.Done()
			bftresult := clientt.Run()
			if clientt.ip == "0" {
				//leader
				participators := []*secp256k1.PublicKey{}
				for index, val := range bftresult.bitmap {
					if val == 1 {
						producer := clientt.Producers[index]
						participators = append(participators, producer.Pubkey)
					}
				}
				msg := bftresult.msg.AsSignMessage()
				sigmaPk := schnorr.CombinePubkeys(participators)

				if !schnorr.Verify(sigmaPk, sha3.Keccak256(msg), bftresult.multiSig.R, bftresult.multiSig.S) {
					t.Error(ErrMultiSig)
				}
			}else{

			}
		}(client)
	}

	group.Wait()
}


func TestBFTTimeOut(t *testing.T) {
	keystore := make([]*secp256k1.PrivateKey, 4)
	produces := make([]consensusTypes.Producer, 4)
	onlinePeers := make(map[string]consensusTypes.IPeerInfo)
	bftClients := make([]*testBFT, 4)
	for i := 0; i < 4; i++ {
		priv, err := secp256k1.GeneratePrivateKey(nil)
		if err != nil {
			i--
			continue
		}
		p := consensusTypes.Producer{priv.PubKey(), strconv.Itoa(i)}
		produces[i] = p
		keystore[i] = priv

		sendor := &testSendor{onlinePeers, bftClients, p.IP}
		bftClient := newTestBFT(keystore[i], produces, sendor, p.IP, onlinePeers)
		bftClients[i] = bftClient
		if i <2 {
			onlinePeers[strconv.Itoa(i)] = &testPeer{p, bftClient}
		}
	}

	group := &sync.WaitGroup{}
	for _, client := range bftClients {
		group.Add(1)
		go func(clientt *testBFT) {
			defer group.Done()
			bftresult := clientt.Run()
			if clientt.ip == "0" {
			 	//leader
			 	if bftresult.err == nil {
			 		t.Error("expect timeout but got success")
				}
			}else if clientt.ip == "1" {
				if bftresult.err == nil {
					t.Error("expect timeout but got success")
				}else{
					if bftresult.err != ErrTimeout {
						t.Errorf("expect timeout err but got %s", bftresult.err)
					}
				}
			}
		}(client)
	}

	group.Wait()
}