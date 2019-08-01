package bft

import (
	"bytes"
	"fmt"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/blockmgr"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/network/p2p"
	p2pService "github.com/drep-project/drep-chain/network/service"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"github.com/drep-project/drep-chain/types"
	"math"
	"math/big"
	"reflect"
	"time"
)

const (
	waitTime = 10 * time.Second
)

type BftConsensus struct {
	CoinBase  crypto.CommonAddress
	PrivKey   *secp256k1.PrivateKey
	curMiner  int
	minMiners int

	BlockMgr     *blockmgr.BlockMgr
	ChainService chain.ChainServiceInterface
	DbService    *database.DatabaseService
	P2pServer    p2pService.P2P

	peersInfo map[string]*consensusTypes.PeerInfo
	WaitTime  time.Duration

	memberMsgPool chan *MsgWrap
	leaderMsgPool chan *MsgWrap

	Producers consensusTypes.ProducerSet
}

func NewBftConsensus(chainService chain.ChainServiceInterface,
	blockMgr *blockmgr.BlockMgr,
	dbService *database.DatabaseService,
	privKey *secp256k1.PrivateKey,
	producer consensusTypes.ProducerSet,
	p2pServer p2pService.P2P,
	peersInfo map[string]*consensusTypes.PeerInfo) *BftConsensus {
	return &BftConsensus{
		CoinBase:      crypto.PubKey2Address(privKey.PubKey()),
		PrivKey:       privKey,
		BlockMgr:      blockMgr,
		ChainService:  chainService,
		DbService:     dbService,
		minMiners:     int(math.Ceil(float64(len(producer)) * 2 / 3)),
		Producers:     producer,
		P2pServer:     p2pServer,
		peersInfo:     peersInfo,
		WaitTime:      waitTime,
		memberMsgPool: make(chan *MsgWrap, 1000),
		leaderMsgPool: make(chan *MsgWrap, 1000),
	}
}

func (bftConsensus *BftConsensus) Run() (*types.Block, error) {
	bftConsensus.minMiners = int(math.Ceil(float64(len(bftConsensus.Producers)) * 2 / 3))
	miners := bftConsensus.collectMemberStatus()
	if len(miners) > 1 {
		isM, isL := bftConsensus.moveToNextMiner(miners)
		if isL {
			//consensusService.leader.UpdateStatus(miners, minMember, consensusService.ChainService.BestChain().Height())
			return bftConsensus.runAsLeader(miners)
		} else if isM {
			//consensusService.member.UpdateStatus(miners, minMember, consensusService.ChainService.BestChain().Height())
			return bftConsensus.runAsMember(miners)
		} else {
			return nil, ErrBFTNotReady
		}
	} else {
		return nil, ErrBFTNotReady
	}
}

func (bftConsensus *BftConsensus) moveToNextMiner(produceInfos []*MemberInfo) (bool, bool) {
	liveMembers := []*MemberInfo{}

	for _, produce := range produceInfos {
		if produce.IsOnline {
			liveMembers = append(liveMembers, produce)
		}
	}
	curentHeight := bftConsensus.ChainService.BestChain().Height()

	liveMinerIndex := int(curentHeight % uint64(len(liveMembers)))
	curMiner := liveMembers[liveMinerIndex]

	for index, produce := range produceInfos {
		if produce.IsOnline {
			if produce.Producer.Pubkey.IsEqual(curMiner.Producer.Pubkey) {
				produce.IsLeader = true
				bftConsensus.curMiner = index
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

func (bftConsensus *BftConsensus) collectMemberStatus() []*MemberInfo {
	produceInfos := make([]*MemberInfo, 0, len(bftConsensus.Producers))
	for _, produce := range bftConsensus.Producers {
		var (
			IsOnline, ok bool
			pi           *consensusTypes.PeerInfo
		)

		isMe := bftConsensus.PrivKey.PubKey().IsEqual(produce.Pubkey)
		if isMe {
			IsOnline = true
		} else {
			//todo  peer获取到的IP地址和配置的ip地址是否相等（nat后是否相等,从tcp原理来看是相等的）
			if pi, ok = bftConsensus.peersInfo[produce.IP]; ok {
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

func (bftConsensus *BftConsensus) runAsMember(miners []*MemberInfo) (block *types.Block, err error) {
	member := NewMember(bftConsensus.PrivKey, bftConsensus.P2pServer, bftConsensus.WaitTime, miners, bftConsensus.minMiners, bftConsensus.ChainService.BestChain().Height(), bftConsensus.memberMsgPool)
	log.Trace("node member is going to process consensus for round 1")
	member.convertor = func(msg []byte) (IConsenMsg, error) {
		block, err = types.BlockFromMessage(msg)
		if err != nil {
			return nil, err
		}

		//faste calc less process time
		calcHash := func(txs []*types.Transaction) {
			for _, tx := range txs {
				tx.TxHash()
			}
		}
		num := 1 + len(block.Data.TxList)/1000
		for i := 0; i < num; i++ {
			if i == num-1 {
				go calcHash(block.Data.TxList[1000*i:])
			} else {
				go calcHash(block.Data.TxList[1000*i : 1000*(i+1)])
			}
		}

		return block, nil
	}
	member.validator = func(msg IConsenMsg) bool {
		block = msg.(*types.Block)
		return bftConsensus.blockVerify(block)
	}
	_, err = member.ProcessConsensus()
	if err != nil {
		return nil, err
	}
	log.Trace("node member finishes consensus for round 1")

	member.Reset()
	log.Trace("node member is going to process consensus for round 2")
	var multiSig *MultiSignature
	member.convertor = func(msg []byte) (IConsenMsg, error) {
		return ResponseWiteRootFromMessage(msg)
	}
	member.validator = func(msg IConsenMsg) bool {
		val := msg.(*ResponseWiteRootMessage)
		multiSig = &val.MultiSignature
		minorAddrs := []crypto.CommonAddress{}
		for index, producer := range bftConsensus.Producers {
			addr := crypto.PubKey2Address(producer.Pubkey)
			if multiSig.Bitmap[index] == 1 && block.Header.LeaderAddress != addr {
				minorAddrs = append(minorAddrs, addr)
			}
		}
		block.Header.MinorAddresses = minorAddrs
		block.Header.StateRoot = val.StateRoot
		proof, err := binary.Marshal(multiSig)
		if err != nil {
			log.Debugf("fial to marshal MultiSig")
			return false
		}
		block.Proof = proof
		return bftConsensus.verifyBlockContent(block)
	}
	_, err = member.ProcessConsensus()
	if err != nil {
		return nil, err
	}
	member.Reset()
	log.Trace("node member finishes consensus for round 2")
	return block, nil
}

//1 leader出块，然后签名并且广播给其他producer,
//2 其他producer收到后，签自己的构建数字签名;然后把签名后的块返回给leader
//3 leader搜集到所有的签名或者返回的签名个数大于producer个数的三分之二后，开始验证签名
//4 leader验证签名通过后，广播此块给所有的Peer
func (bftConsensus *BftConsensus) runAsLeader(miners []*MemberInfo) (block *types.Block, err error) {
	leader := NewLeader(bftConsensus.PrivKey, bftConsensus.P2pServer, bftConsensus.WaitTime, miners, bftConsensus.minMiners, bftConsensus.ChainService.BestChain().Height(), bftConsensus.leaderMsgPool)
	db := bftConsensus.DbService.BeginTransaction(false)
	var gasFee *big.Int
	block, gasFee, err = bftConsensus.BlockMgr.GenerateBlock(db, bftConsensus.CoinBase)
	if err != nil {
		log.WithField("msg", err).Error("generate block fail")
		return nil, err
	}

	log.WithField("Block", block).Trace("node leader is preparing process consensus for round 1")
	err, sig, bitmap := leader.ProcessConsensus(block)
	if err != nil {
		var str = err.Error()
		log.WithField("msg", str).Error("Error occurs")
		return nil, err
	}

	multiSig := &MultiSignature{Sig: *sig, Bitmap: bitmap}
	leader.Reset()
	minorAddrs := []crypto.CommonAddress{}
	for index, producer := range bftConsensus.Producers {
		addr := crypto.PubKey2Address(producer.Pubkey)
		if multiSig.Bitmap[index] == 1 && block.Header.LeaderAddress != addr {
			minorAddrs = append(minorAddrs, crypto.PubKey2Address(producer.Pubkey))
		}
	}
	block.Header.MinorAddresses = minorAddrs
	proof, err := binary.Marshal(multiSig)
	if err != nil {
		log.Debugf("fial to marshal MultiSig")
		return nil, err
	}
	block.Proof = proof
	//Determine reward points
	err = bftConsensus.ChainService.AccumulateRewards(db, block, gasFee)
	if err != nil {
		return nil, err
	}
	db.Commit()
	block.Header.StateRoot = db.GetStateRoot()
	rwMsg := &ResponseWiteRootMessage{*multiSig, block.Header.StateRoot}

	log.Trace("node leader is going to process consensus for round 2")
	err, _, _ = leader.ProcessConsensus(rwMsg)
	if err != nil {
		return nil, err
	}
	log.Trace("node leader finishes process consensus for round 2")
	leader.Reset()
	log.Trace("node leader finishes sending block")
	return block, nil
}

func (bftConsensus *BftConsensus) blockVerify(block *types.Block) bool {
	preBlockHash, err := bftConsensus.ChainService.GetBlockHeaderByHash(&block.Header.PreviousHash)
	if err != nil {
		log.WithField("err", err).Debug("blockVerify GetBlockHeaderByHash")
		return false
	}
	for _, validator := range bftConsensus.ChainService.BlockValidator() {
		if reflect.TypeOf(validator).Elem() != reflect.TypeOf(BlockMultiSigValidator{}) {
			err = validator.VerifyHeader(block.Header, preBlockHash)
			if err != nil {
				log.WithField("err", err).Debug("blockVerify VerifyHeader")
				return false
			}
			err = validator.VerifyBody(block)
			if err != nil {
				log.WithField("err", err).Debug("blockVerify VerifyBody")
				return false
			}
		}
	}
	//TODO need to verify traansaction , a lot of time
	return err == nil
}

func (bftConsensus *BftConsensus) verifyBlockContent(block *types.Block) bool {
	db := bftConsensus.ChainService.GetDatabaseService().BeginTransaction(false)
	multiSigValidator := BlockMultiSigValidator{bftConsensus.Producers}
	if err := multiSigValidator.VerifyBody(block); err != nil {
		return false
	}

	gp := new(chain.GasPool).AddGas(block.Header.GasLimit.Uint64())
	//process transaction
	context := &chain.BlockExecuteContext{
		Db:      db,
		Block:   block,
		Gp:      gp,
		GasUsed: new(big.Int),
		GasFee:  new(big.Int),
	}
	for _, validator := range bftConsensus.ChainService.BlockValidator() {
		err := validator.ExecuteBlock(context)
		if err != nil {
			log.WithField("ExecuteBlock", err).Debug("multySigVerify")
			return false
		}
	}
	err := bftConsensus.ChainService.AccumulateRewards(db, block, context.GasFee)
	if err != nil {
		log.WithField("AccumulateRewards", err).Debug("multySigVerify")
		return false
	}
	db.Commit()
	if block.Header.GasUsed.Cmp(context.GasUsed) == 0 {
		stateRoot := db.GetStateRoot()
		if !bytes.Equal(block.Header.StateRoot, stateRoot) {
			log.Debug("rootcmd root !====")
			return false
		}
	} else {
		log.WithField("gasUsed", context.GasUsed).Debug("multySigVerify")
		return false
	}
	return true
}

func (bftConsensus *BftConsensus) ReceiveMsg(peer *consensusTypes.PeerInfo, rw p2p.MsgReadWriter) error {
	for {
		msg, err := rw.ReadMsg()
		if err != nil {
			log.WithField("Reason", err).Info("consensus receive msg")
			return err
		}

		if msg.Size > MaxMsgSize {
			return ErrMsgSize
		}

		log.WithField("addr", peer.IP()).WithField("code", msg.Code).Debug("Receive setup msg")
		switch msg.Code {
		case MsgTypeSetUp:
			fallthrough
		case MsgTypeChallenge:
			fallthrough
		case MsgTypeFail:
			bftConsensus.memberMsgPool <- &MsgWrap{peer, &msg}
		case MsgTypeCommitment:
			fallthrough
		case MsgTypeResponse:
			bftConsensus.leaderMsgPool <- &MsgWrap{peer, &msg}
		default:
			return fmt.Errorf("consensus unkonw msg type:%d", msg.Code)
		}
	}

	return nil
}

func (bftConsensus *BftConsensus) ChangeTime(interval time.Duration) {
	bftConsensus.WaitTime = interval
}
