package service

import (
	"math"
	"time"
	"encoding/json"

	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/log"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	p2pService "github.com/drep-project/drep-chain/network/service"
	chainService "github.com/drep-project/drep-chain/chain/service"
	consensusTypes "github.com/drep-project/drep-chain/consensus/types"
	accountService "github.com/drep-project/drep-chain/accounts/service"

)

var (
	EnableConsensusFlag = cli.BoolFlag{
		Name:  "enableConsensus",
		Usage: "enable consensus",
	}
)
const (
	blockInterval = time.Second*5
	minWaitTime = time.Millisecond * 500
)

type ConsensusService struct {
	P2pServer     *p2pService.P2pService         `service:"p2p"`
	ChainService  *chainService.ChainService     `service:"chain"`
	DatabaseService  *database.DatabaseService     `service:"database"`
	WalletService *accountService.AccountService `service:"accounts"`

	apis   []app.API
	consensusConfig *consensusTypes.ConsensusConfig

	pubkey *secp256k1.PublicKey
	privkey *secp256k1.PrivateKey
	curMiner int
	leader *Leader
	member *Member
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

func (consensusService *ConsensusService)  P2pMessages() map[int]interface{} {
	return map[int]interface{}{
		consensusTypes.MsgTypeSetUp : consensusTypes.Setup{},
		consensusTypes.MsgTypeCommitment :consensusTypes.Commitment{},
		consensusTypes.MsgTypeChallenge : consensusTypes.Challenge{},
		consensusTypes.MsgTypeResponse : consensusTypes.Response{},
		consensusTypes.MsgTypeFail : consensusTypes.Fail{},
	}
}

func (consensusService *ConsensusService) Init(executeContext *app.ExecuteContext) error {
	consensusService.consensusConfig = &consensusTypes.ConsensusConfig{}
	err := executeContext.UnmashalConfig(consensusService.Name(), consensusService.consensusConfig )
	if err != nil {
		return err
	}

	if executeContext.Cli.IsSet(EnableConsensusFlag.Name) {
		consensusService.consensusConfig.EnableConsensus = executeContext.Cli.GlobalBool(EnableConsensusFlag.Name)
	}
	if !consensusService.consensusConfig.EnableConsensus {
		return nil
	}

	consensusService.pubkey = consensusService.consensusConfig.MyPk
	accountNode, err  := consensusService.WalletService.Wallet.GetAccountByPubkey(consensusService.pubkey)
	if err != nil {
		return err
	}
	consensusService.privkey = accountNode.PrivateKey

	props := actor.FromProducer(func() actor.Actor {
		return consensusService
	})
	pid, err := actor.SpawnNamed(props, "consensus_dbft")
	if err != nil {
		panic(err)
	}

	router :=  consensusService.P2pServer.Router
	chainP2pMessage := consensusService.P2pMessages()
	for msgType, _ := range chainP2pMessage {
		router.RegisterMsgHandler(msgType,pid)
	}
	consensusService.leader = NewLeader(consensusService.pubkey, consensusService.P2pServer)
	consensusService.member = NewMember(consensusService.privkey,consensusService.P2pServer)

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
	return nil
}

func (consensusService *ConsensusService) Start(executeContext *app.ExecuteContext) error {
	if !consensusService.consensusConfig.EnableConsensus {
		return nil
	}
	if !consensusService.isProduce() {
		return nil
	}

	go func() {
		minMember := int(math.Ceil(float64(len(consensusService.consensusConfig.Producers))*2/3)) - 1

		for {
			log.Trace("node start", "Height", consensusService.ChainService.CurrentHeight)
			var block *chainTypes.Block
			var err error
			if consensusService.consensusConfig.ConsensusMode == "solo" {
				block, err = consensusService.runAsSolo()
			} else {
				//TODO a more elegant implementation is needed: select live peer ,and Determine who is the leader
				participants := consensusService.CollectLiveMember()
				if len(participants) > 1 {
					isM, isL := consensusService.MoveToNextMiner(participants)
					if isL {
						consensusService.leader.UpdateStatus(participants, consensusService.curMiner, minMember, consensusService.ChainService.CurrentHeight)
						block, err = consensusService.runAsLeader()
					}else if isM {
						consensusService.member.UpdateStatus(participants, consensusService.curMiner, minMember, consensusService.ChainService.CurrentHeight)
						block, err = consensusService.runAsMember()
					}else{
						// backup node， return directly
						log.Debug("backup node")
						break
					}
				}else{
					err = errors.New("bft node not ready")
					time.Sleep(time.Second*10)
				}
			}
			if err != nil {
				log.Debug("Producer Block Fail", "reason", err.Error())
			}else{
				consensusService.P2pServer.Broadcast(block)
				consensusService.ChainService.ProcessBlock(block)
				log.Info("Block Produced  ", "Height", consensusService.DatabaseService.GetMaxHeight())
			}
			time.Sleep(100) //delay a little time for block deliver
			nextBlockTime, waitSpan :=  consensusService.GetWaitTime()
			log.Debug("Sleep", "nextBlockTime", nextBlockTime, "waitSpan", waitSpan)
			time.Sleep(waitSpan)
			consensusService.OnNewHeightUpdate(consensusService.DatabaseService.GetMaxHeight())
		}
	}()

	return nil
}

func (consensusService *ConsensusService) Stop(executeContext *app.ExecuteContext) error {
	if !consensusService.consensusConfig.EnableConsensus {
		return nil
	}
	return nil
}

func (consensusService *ConsensusService) runAsMember() (*chainTypes.Block, error) {
	consensusService.member.Reset()
	log.Trace("node member is going to process consensus for round 1")
	blockBytes, err := consensusService.member.ProcessConsensus()
	if err != nil {
		return nil, err
	}
	log.Trace("node member finishes consensus for round 1")

	block := &chainTypes.Block{}
	err = json.Unmarshal(blockBytes, block)
	if err != nil {
		return nil, err
	}
	consensusService.member.Reset()
	log.Trace("node member is going to process consensus for round 2")
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
	log.Trace("node member finishes consensus for round 2")
	return block, nil
}

func (consensusService *ConsensusService) runAsLeader() (*chainTypes.Block, error) {
	consensusService.leader.Reset()

	membersPubkey := []*secp256k1.PublicKey{}
	for _, pub := range  consensusService.leader.members {
		membersPubkey = append(membersPubkey, pub.Producer.Public)
	}
	block, err := consensusService.ChainService.GenerateBlock(consensusService.leader.pubkey, membersPubkey)
	if err != nil {
		log.Error("generate block fail", "msg", err )
	}

	log.Trace("node leader is preparing process consensus for round 1", "Block",block)
	msg, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	log.Trace("node leader is going to process consensus for round 1")
	err, sig, bitmap := consensusService.leader.ProcessConsensus(msg)
	if err != nil {
		var str = err.Error()
		log.Error("Error occurs","msg", str)
		return nil, err
	}

	multiSig := &chainTypes.MultiSignature{Sig: *sig, Bitmap: bitmap}
	log.Trace("node leader is preparing process consensus for round 2")
	consensusService.leader.Reset()
	msg, err = json.Marshal(multiSig);
	if err != nil {
		return nil, err
	}
	log.Trace("node leader is going to process consensus for round 2")
	err, _, _ = consensusService.leader.ProcessConsensus(msg)
	if err != nil {
		return nil, err
	}
	log.Trace("node leader finishes process consensus for round 2")
	block.MultiSig = multiSig
	consensusService.leader.Reset()
	log.Trace("node leader finishes sending block")
	return block, nil
}

func (consensusService *ConsensusService) runAsSolo() (*chainTypes.Block, error){
	membersPubkey := []*secp256k1.PublicKey{}
	for _, produce := range  consensusService.consensusConfig.Producers {
		membersPubkey = append(membersPubkey, produce.Public)
	}
	block, _ := consensusService.ChainService.GenerateBlock(consensusService.pubkey, membersPubkey)
	msg, err := json.Marshal(block)
	if err != nil {
		return block, nil
	}

	sig, err := consensusService.privkey.Sign(sha3.Hash256(msg))
	if err != nil {
		log.Error("sign block error")
		return nil, errors.New("sign block error")
	}
	multiSig := &chainTypes.MultiSignature{Sig: *sig, Bitmap: []byte{}}
	block.MultiSig = multiSig
	return block, nil
}

func (consensusService *ConsensusService) isProduce() bool {
	for _, produce := range consensusService.consensusConfig.Producers {
		if produce.Public.IsEqual(consensusService.pubkey){
			return true
		}
	}
	return false
}

func (consensusService *ConsensusService) CollectLiveMember()[]*consensusTypes.MemberInfo{
	liveMembers := []*consensusTypes.MemberInfo{}
	for _, produce := range consensusService.consensusConfig.Producers {
		if consensusService.pubkey.IsEqual(produce.Public) {
			liveMembers = append(liveMembers, &consensusTypes.MemberInfo{
				Producer: produce,
			})  // self
		}else{
			peer := consensusService.P2pServer.GetPeer(produce.Ip)
			if peer != nil {
				liveMembers = append(liveMembers, &consensusTypes.MemberInfo{
					Producer: produce,
					Peer :    peer,
				})
			}
		}
	}
	return liveMembers
}

func (consensusService *ConsensusService) MoveToNextMiner(liveMembers []*consensusTypes.MemberInfo) (bool, bool) {
	consensusService.curMiner = int(consensusService.ChainService.CurrentHeight%int64(len(liveMembers)))

	if liveMembers[consensusService.curMiner].Peer == nil {
		return false, true
	} else{
		return true, false
	}
}

func (consensusService *ConsensusService) OnNewHeightUpdate(height int64) {
	if height > consensusService.ChainService.CurrentHeight {
		consensusService.ChainService.CurrentHeight = height
		log.Info("update new height","Height", height)
	}
}

func (consensusService *ConsensusService) GetMyPubkey() *secp256k1.PublicKey {
	return consensusService.pubkey
}

func (consensusService *ConsensusService) GetWaitTime() (time.Time,time.Duration){
	// max_delay_time +(min_block_interval)*windows = expected_block_interval*windows
	// 6h + 5s*windows = 10s*windows
	// windows = 4320
	lastBlockTime := time.Unix(consensusService.DatabaseService.GetHighestBlock().Header.Timestamp, 0)
	targetTime := lastBlockTime.Add(blockInterval)
	now := time.Now()
	if targetTime.Before(now) {
		return now.Add(time.Millisecond * 500 ), time.Millisecond * 500
	} else{
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
