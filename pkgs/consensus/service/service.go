package service

import (
	"encoding/json"
	"errors"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/app"
	chainService "github.com/drep-project/drep-chain/chain/service"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
	p2pService "github.com/drep-project/drep-chain/network/service"
	walletService "github.com/drep-project/drep-chain/pkgs/wallet/service"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"gopkg.in/urfave/cli.v1"
	"math"
	"time"
)

var (
	EnableConsensusFlag = cli.BoolFlag{
		Name:  "enableConsensus",
		Usage: "enable consensus",
	}
)

const (
	blockInterval = time.Second * 5
	minWaitTime   = time.Millisecond * 500
)

type ConsensusService struct {
	P2pServer       p2pService.P2P                 `service:"p2p"`
	ChainService    *chainService.ChainService     `service:"chain"`
	DatabaseService *database.DatabaseService      `service:"database"`
	WalletService   *walletService.AccountService `service:"accounts"`

	apis   []app.API
	Config *consensusTypes.ConsensusConfig

	signPubkey         *secp256k1.PublicKey
	signPrivkey        *secp256k1.PrivateKey

	curMiner           int
	leader             *Leader
	member             *Member
	syncBlockEventSub  event.Subscription
	syncBlockEventChan chan event.SyncBlockEvent

	//During the process of synchronizing blocks, the miner stopped mining
	pauseForSync bool
	start        bool

	quit chan struct{}
}

func (consensusService *ConsensusService) Name() string {
	return "consensus"
}

func (consensusService *ConsensusService) Api() []app.API {
	return consensusService.apis
}

func (consensusService *ConsensusService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{EnableConsensusFlag}
}

func (consensusService *ConsensusService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{
		consensusTypes.MsgTypeSetUp:      consensusTypes.Setup{},
		consensusTypes.MsgTypeCommitment: consensusTypes.Commitment{},
		consensusTypes.MsgTypeChallenge:  consensusTypes.Challenge{},
		consensusTypes.MsgTypeResponse:   consensusTypes.Response{},
		consensusTypes.MsgTypeFail:       consensusTypes.Fail{},
	}
}

func (consensusService *ConsensusService) Init(executeContext *app.ExecuteContext) error {
	consensusService.Config = &consensusTypes.ConsensusConfig{}
	err := executeContext.UnmashalConfig(consensusService.Name(), consensusService.Config)
	if err != nil {
		return err
	}

	if executeContext.Cli.IsSet(EnableConsensusFlag.Name) {
		consensusService.Config.EnableConsensus = executeContext.Cli.GlobalBool(EnableConsensusFlag.Name)
	}
	if !consensusService.Config.EnableConsensus {
		return nil
	}

	consensusService.signPubkey = consensusService.GetSignPubkey(consensusService.Config.Me)
	if consensusService.signPubkey == nil {
		return  errors.New("consensus config error")
	}
	consensusService.signPrivkey, err = consensusService.WalletService.Wallet.DumpPrivateKey(consensusService.signPubkey)
	if err != nil {
		dlog.Error("consensusService", "init err", err, "pubkey", string(consensusService.pubkey.Serialize()))
		return err
	}

	props := actor.FromProducer(func() actor.Actor {
		return consensusService
	})
	pid, err := actor.SpawnNamed(props, "consensus_dbft")
	if err != nil {
		panic(err)
	}

	router := consensusService.P2pServer.GetRouter()
	chainP2pMessage := consensusService.P2pMessages()
	for msgType, _ := range chainP2pMessage {
		router.RegisterMsgHandler(msgType, pid)
	}
	consensusService.leader = NewLeader(consensusService.signPubkey, consensusService.P2pServer)
	consensusService.member = NewMember(consensusService.signPrivkey, consensusService.P2pServer)
	consensusService.syncBlockEventChan = make(chan event.SyncBlockEvent)
	consensusService.syncBlockEventSub = consensusService.ChainService.SubscribeSyncBlockEvent(consensusService.syncBlockEventChan)
	consensusService.quit = make(chan struct{})

	consensusService.apis = []app.API{
		app.API{
			Namespace: "consensus",
			Version:   "1.0",
			Service: &ConsensusApi{
				consensusService: consensusService,
			},
			Public: true,
		},
	}

	go consensusService.handlerEvent()

	return nil
}

func (consensusService *ConsensusService) handlerEvent() {
	for {
		select {
		case e := <-consensusService.syncBlockEventChan:
			if e.EventType == event.StartSyncBlock {
				consensusService.pauseForSync = true
				dlog.Info("Start Sync Blcok")
			} else {
				consensusService.pauseForSync = false
				dlog.Info("Stop Sync Blcok")
			}
		case <-consensusService.quit:
			return
		}
	}
}

func (consensusService *ConsensusService) Start(executeContext *app.ExecuteContext) error {
	if !consensusService.Config.EnableConsensus {
		return nil
	}
	if !consensusService.isProduce() {
		return nil
	}

	consensusService.start = true

	go func() {
		minMember := int(math.Ceil(float64(len(consensusService.Config.Producers))*2/3)) - 1

		select {
		case <-consensusService.quit:
			return
		default:
			for {
				if consensusService.pauseForSync {
					time.Sleep(time.Millisecond * 500)
					continue
				}
				dlog.Trace("node start", "Height", consensusService.ChainService.BestChain.Height())
				var block *chainTypes.Block
				var err error
				var isM bool
				var isL bool
				if consensusService.Config.ConsensusMode == "solo" {
					block, err = consensusService.runAsSolo()
					isL = true
				} else if consensusService.Config.ConsensusMode == "bft" {
					//TODO a more elegant implementation is needed: select live peer ,and Determine who is the leader
					miners := consensusService.collectMemberStatus()
					if len(miners) > 1 {
						isM, isL = consensusService.moveToNextMiner(miners)
						if isL {
							consensusService.leader.UpdateStatus(miners, minMember, consensusService.ChainService.BestChain.Height())
							block, err = consensusService.runAsLeader()
						} else if isM {
							consensusService.member.UpdateStatus(miners, minMember, consensusService.ChainService.BestChain.Height())
							block, err = consensusService.runAsMember()
						} else {
							// backup node， return directly
							dlog.Debug("Backup Node")
							break
						}
					} else {
						err = errors.New("BFT node not ready")
						time.Sleep(time.Second * 10)
					}
				} else {
					break
				}
				if err != nil {
					dlog.Debug("Producer Block Fail", "Reason", err.Error())
				} else {
					if isL {
						consensusService.P2pServer.Broadcast(block)
						consensusService.ChainService.ProcessBlock(block)
						dlog.Info("Submit Block ", "Height", consensusService.ChainService.BestChain.Height(), "txs:", block.Data.TxCount)
					}
				}
				time.Sleep(time.Duration(500) * time.Millisecond) //delay a little time for block deliver
				nextBlockTime, waitSpan := consensusService.getWaitTime()
				dlog.Debug("Sleep", "nextBlockTime", nextBlockTime, "waitSpan", waitSpan)
				time.Sleep(waitSpan)
			}
		}
	}()

	return nil
}

func (consensusService *ConsensusService) Stop(executeContext *app.ExecuteContext) error {
	if consensusService.Config != nil && !consensusService.Config.EnableConsensus {
		return nil
	}
	close(consensusService.quit)
	consensusService.syncBlockEventSub.Unsubscribe()
	return nil
}

func (consensusService *ConsensusService) runAsMember() (*chainTypes.Block, error) {
	consensusService.member.Reset()
	dlog.Trace("node member is going to process consensus for round 1")
	blockBytes, err := consensusService.member.ProcessConsensus()
	if err != nil {
		return nil, err
	}
	dlog.Trace("node member finishes consensus for round 1")

	block := &chainTypes.Block{}
	err = json.Unmarshal(blockBytes, block)
	if err != nil {
		return nil, err
	}
	consensusService.member.Reset()
	dlog.Trace("node member is going to process consensus for round 2")
	multiSigBytes, err := consensusService.member.ProcessConsensus()
	if err != nil {
		return nil, err
	}
	multiSig := &chainTypes.MultiSignature{}
	err = json.Unmarshal(multiSigBytes, multiSig)
	if err != nil {
		return nil, err
	}
	block.MultiSig = multiSig
	//check multiSig

	/*
	sigmaPubKey := schnorr.CombinePubkeys(pubkeys)
	isValid :=  schnorr.Verify(sigmaPubKey, sha3.Hash256(blockBytes), multiSig.Sig.R, multiSig.Sig.S)
	if !isValid {
		return nil, errors.New("signature not correct")
	}
	*/
	consensusService.leader.Reset()
	dlog.Trace("node member finishes consensus for round 2")
	return block, nil
}

//1 leader出块，然后签名并且广播给其他producer,
//2 其他producer收到后，签自己的构建数字签名;然后把签名后的块返回给leader
//3 leader搜集到所有的签名或者返回的签名个数大于producer个数的三分之二后，开始验证签名
//4 leader验证签名通过后，广播此块给所有的Peer
func (consensusService *ConsensusService) runAsLeader() (*chainTypes.Block, error) {
	consensusService.leader.Reset()

	membersAccounts := []string{}
	for _, pub := range consensusService.leader.members {
		membersAccounts = append(membersAccounts, pub.Producer.Account)
	}
	block, err := consensusService.ChainService.GenerateBlock(consensusService.Config.Me, membersAccounts)
	if err != nil {
		dlog.Error("generate block fail", "msg", err)
	}

	dlog.Trace("node leader is preparing process consensus for round 1", "Block", block)
	msg, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	dlog.Trace("node leader is going to process consensus for round 1")
	err, sig, bitmap := consensusService.leader.ProcessConsensus(msg)
	if err != nil {
		var str = err.Error()
		dlog.Error("Error occurs", "msg", str)
		return nil, err
	}

	multiSig := &chainTypes.MultiSignature{Sig: *sig, Bitmap: bitmap}
	dlog.Trace("node leader is preparing process consensus for round 2")
	consensusService.leader.Reset()
	msg, err = json.Marshal(multiSig)
	if err != nil {
		return nil, err
	}
	dlog.Trace("node leader is going to process consensus for round 2")
	err, _, _ = consensusService.leader.ProcessConsensus(msg)
	if err != nil {
		return nil, err
	}
	dlog.Trace("node leader finishes process consensus for round 2")
	block.MultiSig = multiSig
	consensusService.leader.Reset()
	dlog.Trace("node leader finishes sending block")
	return block, nil
}

func (consensusService *ConsensusService) runAsSolo() (*chainTypes.Block, error) {
	membersAccounts := []string{}
	for _, produce := range consensusService.Config.Producers {
		membersAccounts = append(membersAccounts, produce.Account)
	}
	block, _ := consensusService.ChainService.GenerateBlock(consensusService.Config.Me, membersAccounts)
	msg, err := json.Marshal(block)
	if err != nil {
		return block, nil
	}

	sig, err := consensusService.signPrivkey.Sign(sha3.Hash256(msg))
	if err != nil {
		dlog.Error("sign block error")
		return nil, errors.New("sign block error")
	}
	multiSig := &chainTypes.MultiSignature{Sig: *sig, Bitmap: []byte{}}
	block.MultiSig = multiSig
	return block, nil
}

func (consensusService *ConsensusService) isProduce() bool {
	for _, produce := range consensusService.Config.Producers {
		if produce.SignPubkey.IsEqual(consensusService.signPubkey) {
			return true
		}
	}
	return false
}

func (consensusService *ConsensusService) collectMemberStatus() []*consensusTypes.MemberInfo {
	produceInfos := make([]*consensusTypes.MemberInfo, len(consensusService.Config.Producers))
	for i, produce := range consensusService.Config.Producers {
		peer := consensusService.P2pServer.GetPeer(produce.Ip)
		isOnLine := peer != nil
		isMe := consensusService.pubkey.IsEqual(produce.Public)
		if isMe {
			isOnLine = true
		}
		produceInfos[i] = &consensusTypes.MemberInfo{
			Producer: produce,
			Peer:     peer,
			IsMe:     isMe,
			IsOnline: isOnLine,
		}
	}
	return produceInfos
}

func (consensusService *ConsensusService) moveToNextMiner(produceInfos []*consensusTypes.MemberInfo) (bool, bool) {
	liveMembers := []*consensusTypes.MemberInfo{}

	for _, produce := range produceInfos {
		if produce.IsOnline {
			liveMembers = append(liveMembers, produce)
		}
	}
	curentHeight := consensusService.ChainService.BestChain.Height()
	liveMinerIndex := int(curentHeight % int64(len(liveMembers)))
	curMiner := liveMembers[liveMinerIndex]

	for index, produce := range produceInfos {
		if produce.IsOnline {
			if produce.Producer.Public.IsEqual(curMiner.Producer.Public) {
				produce.IsLeader = true
				consensusService.curMiner = index
			} else {
				produce.IsLeader = false
			}
		}
	}

	if curMiner.Peer == nil {
		return false, true
	} else {
		return true, false
	}
}

func (consensusService *ConsensusService) GetSignPubkey(accountName string) *secp256k1.PublicKey {
	for _, producer := range consensusService.Config.Producers {
		if producer.Account == accountName {
			return &producer.SignPubkey
		}
	}
	return nil
}

func (consensusService *ConsensusService) GetWaitTime() (time.Time, time.Duration) {
	// max_delay_time +(min_block_interval)*windows = expected_block_interval*windows
	// 6h + 5s*windows = 10s*windows
	// windows = 4320

	lastBlockTime := time.Unix(consensusService.ChainService.BestChain.Tip().TimeStamp, 0)
	targetTime := lastBlockTime.Add(blockInterval)
	now := time.Now()
	if targetTime.Before(now) {
		return now.Add(time.Millisecond * 500), time.Millisecond * 500
	} else {
		return targetTime, targetTime.Sub(now)
	}
	/*
     window := int64(4320)
     endBlock := consensusService.DatabaseService.GetHighestBlock().Header
     if endBlock.Height < window {
		 lastBlockTime := time.Unix(consensusService.DatabaseService.GetHighestBlock().Header.Timestamp, 0)
		 span := time.Now().Sub(lastBlockTime)
		 if span > blockInterval {
			 span = 0
		 } else {
			 span = blockInterval - span
		 }
		 return span
	 }else{
	 	//wait for test
		 startHeight := endBlock.Height - window
		 if startHeight <0 {
			 startHeight = int64(0)
		 }
		 startBlock :=consensusService.DatabaseService.GetBlock(startHeight).Header

		 xx := window * 10 -(time.Unix(startBlock.Timestamp,0).Sub(time.Unix(endBlock.Timestamp,0))).Seconds()

		 span := time.Unix(startBlock.Timestamp,0).Sub(time.Unix(endBlock.Timestamp,0))  //window time
		 avgSpan := span.Nanoseconds()/window
		 return time.Duration(avgSpan) * time.Nanosecond
	 }
	*/
}
