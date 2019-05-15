package service

import (
	"errors"
	"fmt"
	"github.com/drep-project/binary"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/app"
	blockMgrService "github.com/drep-project/drep-chain/chain/service/blockmgr"
	chainService "github.com/drep-project/drep-chain/chain/service/chainservice"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/network/p2p"
	p2pService "github.com/drep-project/drep-chain/network/service"
	accountService "github.com/drep-project/drep-chain/pkgs/accounts/service"
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
	P2pServer    p2pService.P2P             `service:"p2p"`
	ChainService *chainService.ChainService `service:"chain"`
	BlockMgr     *blockMgrService.BlockMgr  `service:"blockmgr"`

	DatabaseService *database.DatabaseService      `service:"database"`
	WalletService   *accountService.AccountService `service:"accounts"`

	apis   []app.API
	Config *consensusTypes.ConsensusConfig

	pubkey             *secp256k1.PublicKey
	privkey            *secp256k1.PrivateKey
	curMiner           int
	leader             *Leader
	member             *Member
	syncBlockEventSub  event.Subscription
	syncBlockEventChan chan event.SyncBlockEvent

	//During the process of synchronizing blocks, the miner stopped mining
	pauseForSync bool
	start        bool
	peersInfo    map[string]*consensusTypes.PeerInfo
	producers    map[string]*secp256k1.PublicKey

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

func (consensusService *ConsensusService) Init(executeContext *app.ExecuteContext) error {
	if consensusService.ChainService == nil {
		return fmt.Errorf("chainService not init")
	}
	consensusService.producers = make(map[string]*secp256k1.PublicKey)
	for _, producer := range consensusService.ChainService.Config.Producers {
		consensusService.producers[producer.IP] = producer.Pubkey
	}

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

	consensusService.pubkey = consensusService.Config.MyPk
	accountNode, err := consensusService.WalletService.Wallet.GetAccountByPubkey(consensusService.pubkey)
	if err != nil {
		dlog.Error("consensusService", "init err", err, "pubkey", string(consensusService.pubkey.Serialize()))
		return err
	}
	consensusService.privkey = accountNode.PrivateKey
	consensusService.peersInfo = make(map[string]*consensusTypes.PeerInfo)

	consensusService.P2pServer.AddProtocols([]p2p.Protocol{
		p2p.Protocol{
			Name:   "consensusService",
			Length: consensusTypes.NumberOfMsg,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				if _, ok := consensusService.producers[peer.IP()]; ok {
					pi := consensusTypes.NewPeerInfo(peer, rw)
					consensusService.peersInfo[peer.IP()] = pi
					defer delete(consensusService.peersInfo, peer.IP())
					return consensusService.receiveMsg(pi, rw)
				}
				//非骨干节点，不启动共识相关处理
				return nil
			},
		},
	})

	consensusService.leader = NewLeader(consensusService.privkey, consensusService.P2pServer)
	consensusService.member = NewMember(consensusService.privkey, consensusService.P2pServer)
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
		minMember := int(math.Ceil(float64(len(consensusService.producers)) * 2 / 3))

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
						err =  ErrBFTNotReady
						time.Sleep(time.Second * 10)
					}
				} else {
					break
				}
				if err != nil {
					dlog.Debug("Producer Block Fail", "Reason", err.Error())
				} else {
					_, _, err := consensusService.ChainService.ProcessBlock(block)
					if err == nil{
						consensusService.BlockMgr.BroadcastBlock(chainTypes.MsgTypeBlock, block, true)
					}
					dlog.Info("Submit Block ", "Height", consensusService.ChainService.BestChain.Height(), "txs:", block.Data.TxCount, "err", err)
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
	if consensusService.Config == nil || !consensusService.Config.EnableConsensus {
		return nil
	}

	close(consensusService.quit)
	consensusService.syncBlockEventSub.Unsubscribe()
	return nil
}

func (consensusService *ConsensusService) runAsMember() (*chainTypes.Block, error) {
	consensusService.member.Reset()

	dlog.Trace("node member is going to process consensus for round 1")
	block := &chainTypes.Block{}
	consensusService.member.validator = func(msg []byte) bool {
		err := binary.Unmarshal(msg, block)
		if err != nil {
			return false
		}
		return consensusService.blockVerify(block)
	}
	_, err := consensusService.member.ProcessConsensus()
	if err != nil {
		return nil, err
	}
	dlog.Trace("node member finishes consensus for round 1")

	consensusService.member.Reset()
	multiSig := &chainTypes.MultiSignature{}
	consensusService.member.validator = func(multiSigBytes []byte) bool {
		err := binary.Unmarshal(multiSigBytes, multiSig)
		if err != nil {
			return false
		}
		minorPubkeys := []secp256k1.PublicKey{}
		for index, producer := range consensusService.ChainService.Config.Producers {
			if multiSig.Bitmap[index] == 1 {
				minorPubkeys = append(minorPubkeys, *producer.Pubkey)
			}
		}
		block.Header.MinorPubKeys = minorPubkeys
		block.MultiSig = multiSig
		return consensusService.multySigVerify(block)
	}
	dlog.Trace("node member is going to process consensus for round 2")
	_, err = consensusService.member.ProcessConsensus()
	if err != nil {
		return nil, err
	}
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
	db := consensusService.DatabaseService.BeginTransaction()
	defer db.Discard()
	block, gasFee ,err := consensusService.BlockMgr.GenerateBlock(db, consensusService.leader.pubkey)
	if err != nil {
		dlog.Error("generate block fail", "msg", err)
		return nil, err
	}

	dlog.Trace("node leader is preparing process consensus for round 1", "Block", block)
	err, sig, bitmap := consensusService.leader.ProcessConsensus(block.ToMessage())
	if err != nil {
		var str = err.Error()
		dlog.Error("Error occurs", "msg", str)
		return nil, err
	}

	multiSig := &chainTypes.MultiSignature{Sig: *sig, Bitmap: bitmap}
	consensusService.leader.Reset()
	msg, err := binary.Marshal(multiSig)
	if err != nil {
		return nil, err
	}
	minorPubkeys := []secp256k1.PublicKey{}
	for index, producer := range consensusService.ChainService.Config.Producers {
		if multiSig.Bitmap[index] == 1 {
			minorPubkeys = append(minorPubkeys, *producer.Pubkey)
		}
	}
	block.Header.MinorPubKeys = minorPubkeys
	block.MultiSig = multiSig
	//Determine reward points
    err = consensusService.ChainService.AccumulateRewards(db, block, gasFee)
	if err != nil {
		return nil, err
	}
	block.Header.StateRoot = db.GetStateRoot()
	dlog.Trace("node leader is going to process consensus for round 2")
	err, _, _ = consensusService.leader.ProcessConsensus(msg)
	if err != nil {
		return nil, err
	}
	dlog.Trace("node leader finishes process consensus for round 2")
	consensusService.leader.Reset()
	dlog.Trace("node leader finishes sending block")
	return block, nil
}

func (consensusService *ConsensusService) runAsSolo() (*chainTypes.Block, error) {
	db := consensusService.DatabaseService.BeginTransaction()
	defer db.Discard()
	block, gasFee, err := consensusService.BlockMgr.GenerateBlock(db, consensusService.pubkey)
	if err != nil {
		return nil, err
	}
	msg, err := binary.Marshal(block)
	if err != nil {
		return block, nil
	}

	sig, err := consensusService.privkey.Sign(sha3.Keccak256(msg))
	if err != nil {
		dlog.Error("sign block error")
		return nil, errors.New("sign block error")
	}
	multiSig := &chainTypes.MultiSignature{Sig: *sig, Bitmap: []byte{1}}
	block.MultiSig = multiSig
	err = consensusService.ChainService.AccumulateRewards(db, block, gasFee)
	if err != nil {
		return nil, err
	}
	block.Header.StateRoot = db.GetStateRoot()
	return block, nil
}

func (consensusService *ConsensusService) isProduce() bool {
	for _, pubkey := range consensusService.producers {
		if pubkey.IsEqual(consensusService.pubkey) {
			return true
		}
	}
	return false
}

func (consensusService *ConsensusService) collectMemberStatus() []*consensusTypes.MemberInfo {
	produceInfos := make([]*consensusTypes.MemberInfo, 0, len(consensusService.producers))
	for ip, pubkey := range consensusService.producers {
		var (
			IsOnline, ok bool
			pi           *consensusTypes.PeerInfo
		)

		isMe := consensusService.pubkey.IsEqual(pubkey)
		if isMe {
			IsOnline = true
		} else {
			//todo  peer获取到的IP地址和配置的ip地址是否相等（nat后是否相等,从tcp原理来看是相等的）
			if pi, ok = consensusService.peersInfo[ip]; ok {
				IsOnline = true
			}
		}

		produceInfos = append(produceInfos, &consensusTypes.MemberInfo{
			Producer: &consensusTypes.Producer{Public: pubkey, Ip: ip},
			Peer:     pi,
			IsMe:     isMe,
			IsOnline: IsOnline,
		})
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
	liveMinerIndex := int(curentHeight % uint64(len(liveMembers)))
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

//本函数只能广播本模块定义的消息
func (consensusService *ConsensusService) BroadcastConsensusMsg(msgType int, msg interface{}) {
	go func() {
		for _, peer := range consensusService.peersInfo {
			consensusService.P2pServer.Send(peer.GetMsgRW(), uint64(msgType), msg)
		}
	}()
}

func (consensusService *ConsensusService) getWaitTime() (time.Time, time.Duration) {
	// max_delay_time +(min_block_interval)*windows = expected_block_interval*windows
	// 6h + 5s*windows = 10s*windows
	// windows = 4320

	lastBlockTime := time.Unix(int64(consensusService.ChainService.BestChain.Tip().TimeStamp), 0)
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

func (consensusService *ConsensusService) blockVerify(block *chainTypes.Block) bool {
	preBlockHash, err := consensusService.ChainService.GetBlockHeaderByHash(&block.Header.PreviousHash)
	if err != nil {
		return false
	}
	err = consensusService.ChainService.BlockValidator.VerifyHeader(block.Header, preBlockHash)
	if err != nil {
		return false
	}
	err = consensusService.ChainService.BlockValidator.VerifyBody(block)
	if err != nil {
		return false
	}
	//TODO need to verify traansaction , a lot of time
	return err == nil
}

func (consensusService *ConsensusService) multySigVerify(block *chainTypes.Block) bool {
	return consensusService.ChainService.BlockValidator.VerifyMultiSig(block, consensusService.ChainService.Config.SkipCheckMutiSig || false)
}
