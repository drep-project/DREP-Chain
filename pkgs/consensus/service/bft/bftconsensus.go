package bft

import (
	"bytes"
	"github.com/drep-project/DREP-Chain/chain/store"
	"math"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/drep-project/binary"
	"github.com/drep-project/DREP-Chain/blockmgr"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/common/event"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/database"

	consensusTypes "github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	"github.com/drep-project/DREP-Chain/types"
)

const (
	waitTime = 10 * time.Second
)

type BftConsensus struct {
	CoinBase crypto.CommonAddress
	PrivKey  *secp256k1.PrivateKey
	config   *BftConfig
	curMiner int

	BlockGenerator blockmgr.IBlockBlockGenerator

	ChainService chain.ChainServiceInterface
	DbService    *database.DatabaseService
	sender       Sender

	peerLock   sync.RWMutex
	onLinePeer map[string]consensusTypes.IPeerInfo
	WaitTime   time.Duration

	memberMsgPool chan *MsgWrap
	leaderMsgPool chan *MsgWrap

	addPeerChan    chan *consensusTypes.PeerInfo
	removePeerChan chan *consensusTypes.PeerInfo

	producer []*Producer
}

func NewBftConsensus(
	chainService chain.ChainServiceInterface,
	blockGenerator blockmgr.IBlockBlockGenerator,
	dbService *database.DatabaseService,
	sener Sender,
	config *BftConfig,
	addPeer, removePeer *event.Feed) *BftConsensus {
	addPeerChan := make(chan *consensusTypes.PeerInfo)
	removePeerChan := make(chan *consensusTypes.PeerInfo)
	addPeer.Subscribe(addPeerChan)
	removePeer.Subscribe(removePeerChan)
	return &BftConsensus{
		BlockGenerator: blockGenerator,
		ChainService:   chainService,
		config:         config,
		DbService:      dbService,
		sender:         sener,
		onLinePeer:     map[string]consensusTypes.IPeerInfo{},
		WaitTime:       waitTime,
		addPeerChan:    addPeerChan,
		removePeerChan: removePeerChan,
		memberMsgPool : make(chan *MsgWrap, 1000),
		leaderMsgPool : make(chan *MsgWrap, 1000),
	}
}

func (bftConsensus *BftConsensus) GetProducers(height uint64, topN int) ([]*Producer, error) {
	newEpoch := height % uint64(bftConsensus.config.ChangeInterval)
	if bftConsensus.producer == nil || newEpoch == 0 {
		height = height - newEpoch
		block, err := bftConsensus.ChainService.GetBlockByHeight(height)
		if err != nil {
			return nil, err
		}
		trie, err := store.TrieStoreFromStore(bftConsensus.DbService.LevelDb(), block.Header.StateRoot)
		if err != nil {
			return nil, err
		}
		producers := GetCandidates(trie, topN)
		bftConsensus.producer = producers
		return producers, nil
	} else {
		return bftConsensus.producer, nil
	}
}

func (bftConsensus *BftConsensus) Run(privKey *secp256k1.PrivateKey) (*types.Block, error) {
	bftConsensus.CoinBase = crypto.PubkeyToAddress(privKey.PubKey())
	bftConsensus.PrivKey = privKey
	go bftConsensus.processPeers()
	producers, err := bftConsensus.GetProducers(bftConsensus.ChainService.BestChain().Tip().Height, bftConsensus.config.ProducerNum)
	if err != nil {
		return nil, err
	}
	found := false
	for _, p := range producers {
		if bytes.Equal(p.Pubkey.Serialize(), bftConsensus.config.MyPk.Serialize()) {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrNotMyTurn
	}
	minMiners := int(math.Ceil(float64(len(producers)) * 2 / 3))
	miners := bftConsensus.collectMemberStatus(producers)
	//print miners status
	str := "-----------------------------------\n"
	for _, m := range miners {
		str += "|	"+m.Producer.Node.IP().String() + "|	"+ strconv.FormatBool(m.IsOnline)  +"	|\n"
	}
	fmt.Println(str)

	if len(miners) > 1 {
		isM, isL, err := bftConsensus.moveToNextMiner(miners)
		if err != nil {
			return nil, err
		}
		if isL {
			return bftConsensus.runAsLeader(producers, miners, minMiners)
		} else if isM {
			return bftConsensus.runAsMember(miners, minMiners)
		} else {
			return nil, ErrBFTNotReady
		}
	} else {
		return nil, ErrBFTNotReady
	}
}

func (bftConsensus *BftConsensus) processPeers() {
	for {
		select {
		case addPeer := <-bftConsensus.addPeerChan:
			bftConsensus.peerLock.Lock()
			bftConsensus.onLinePeer[addPeer.IP()] = addPeer
			bftConsensus.peerLock.Unlock()
		case removePeer := <-bftConsensus.removePeerChan:
			bftConsensus.peerLock.Lock()
			delete(bftConsensus.onLinePeer, removePeer.IP())
			bftConsensus.peerLock.Unlock()
		}
	}
}

func (bftConsensus *BftConsensus) moveToNextMiner(produceInfos []*MemberInfo) (bool, bool, error) {
	liveMembers := []*MemberInfo{}
	for _, produce := range produceInfos {
		if produce.IsOnline {
			liveMembers = append(liveMembers, produce)
		}
	}
	curentHeight := bftConsensus.ChainService.BestChain().Height()
	if uint64(len(liveMembers)) == 0 {
		return false, false, ErrBFTNotReady
	}
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

	if curMiner.IsMe {
		return false, true, nil
	} else {
		return true, false, nil
	}
}

func (bftConsensus *BftConsensus) collectMemberStatus(producers []*Producer) []*MemberInfo {
	produceInfos := make([]*MemberInfo, 0, len(producers))
	for _, produce := range producers {
		var (
			IsOnline, ok bool
			pi           consensusTypes.IPeerInfo
		)

		isMe := bftConsensus.PrivKey.PubKey().IsEqual(produce.Pubkey)
		if isMe {
			IsOnline = true
		} else {
			//todo  peer获取到的IP地址和配置的ip地址是否相等（nat后是否相等,从tcp原理来看是相等的）
			bftConsensus.peerLock.RLock()
			if pi, ok = bftConsensus.onLinePeer[produce.Node.IP().String()]; ok {
				IsOnline = true
			}
			bftConsensus.peerLock.RUnlock()
		}

		produceInfos = append(produceInfos, &MemberInfo{
			Producer: &Producer{Pubkey: produce.Pubkey, Node: produce.Node},
			Peer:     pi,
			IsMe:     isMe,
			IsOnline: IsOnline,
		})
	}
	return produceInfos
}

func (bftConsensus *BftConsensus) runAsMember(miners []*MemberInfo, minMiners int) (block *types.Block, err error) {
	member := NewMember(bftConsensus.PrivKey, bftConsensus.sender, bftConsensus.WaitTime, miners, minMiners, bftConsensus.ChainService.BestChain().Height(), bftConsensus.memberMsgPool)
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
	member.validator = func(msg IConsenMsg) error {
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
		return CompletedBlockFromMessage(msg)
	}

	member.validator = func(msg IConsenMsg) error {
		val := msg.(*CompletedBlockMessage)
		multiSig = &val.MultiSignature
		block.Header.StateRoot = val.StateRoot
		multiSigBytes, err := binary.Marshal(multiSig)
		if err != nil {
			log.Error("fail to marshal MultiSig")
			return err
		}
		log.WithField("bitmap", multiSig.Bitmap).Info("member receive participant bitmap")
		block.Proof = types.Proof{consensusTypes.Pbft, multiSigBytes}
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
func (bftConsensus *BftConsensus) runAsLeader(producers ProducerSet, miners []*MemberInfo, minMiners int) (block *types.Block, err error) {
	leader := NewLeader(
		bftConsensus.PrivKey,
		bftConsensus.sender,
		bftConsensus.WaitTime,
		miners,
		minMiners,
		bftConsensus.ChainService.BestChain().Height(),
		bftConsensus.leaderMsgPool)
	defer leader.Close()
	trieStore, err := store.TrieStoreFromStore(bftConsensus.DbService.LevelDb(), bftConsensus.ChainService.BestChain().Tip().StateRoot)
	if err != nil {
		return nil, err
	}
	var gasFee *big.Int
	block, gasFee, err = bftConsensus.BlockGenerator.GenerateTemplate(trieStore, bftConsensus.CoinBase)
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

	leader.Reset()
	multiSig := newMultiSignature(*sig, bftConsensus.curMiner, bitmap)
	multiSigBytes, err := binary.Marshal(multiSig)
	if err != nil {
		log.Debugf("fial to marshal MultiSig")
		return nil, err
	}
	log.WithField("bitmap", multiSig.Bitmap).Info("participant bitmap")
	//Determine reward points
	block.Proof = types.Proof{consensusTypes.Pbft, multiSigBytes}
	calculator := NewRewardCalculator(trieStore, multiSig, producers, gasFee, block.Header.Height)
	err = calculator.AccumulateRewards()
	if err != nil {
		return nil, err
	}

	block.Header.StateRoot = trieStore.GetStateRoot()
	rwMsg := &CompletedBlockMessage{*multiSig, block.Header.StateRoot}

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

func (bftConsensus *BftConsensus) blockVerify(block *types.Block) error {
	preBlockHash, err := bftConsensus.ChainService.GetBlockHeaderByHash(&block.Header.PreviousHash)
	if err != nil {
		return err
	}
	for _, validator := range bftConsensus.ChainService.BlockValidator() {
		if reflect.TypeOf(validator).Elem() != reflect.TypeOf(BlockMultiSigValidator{}) {
			err = validator.VerifyHeader(block.Header, preBlockHash)
			if err != nil {
				return err
			}
			err = validator.VerifyBody(block)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (bftConsensus *BftConsensus) verifyBlockContent(block *types.Block) error {
	parent, err := bftConsensus.ChainService.GetBlockHeaderByHeight(block.Header.Height - 1)
	if err != nil {
		return err
	}
	dbstore := &chain.ChainStore{bftConsensus.DbService.LevelDb()}
	trieStore, err := store.TrieStoreFromStore(bftConsensus.DbService.LevelDb(), parent.StateRoot)
	if err != nil {
		return err
	}
	multiSigValidator := BlockMultiSigValidator{bftConsensus.GetProducers, bftConsensus.ChainService.GetBlockByHash, bftConsensus.config.ProducerNum}
	if err := multiSigValidator.VerifyBody(block); err != nil {
		return err
	}

	gp := new(chain.GasPool).AddGas(block.Header.GasLimit.Uint64())
	//process transaction
	context := chain.NewBlockExecuteContext(trieStore, gp, dbstore, block)
	for _, validator := range bftConsensus.ChainService.BlockValidator() {
		err := validator.ExecuteBlock(context)
		if err != nil {
			return err
		}
	}
	multiSig := &MultiSignature{}
	err = binary.Unmarshal(block.Proof.Evidence, multiSig)
	if err != nil {
		return err
	}

	stateRoot := trieStore.GetStateRoot()
	if block.Header.GasUsed.Cmp(context.GasUsed) == 0 {
		if !bytes.Equal(block.Header.StateRoot, stateRoot) {
			if !trieStore.RecoverTrie(bftConsensus.ChainService.GetCurrentHeader().StateRoot) {
				log.Error("root not equal and recover trie err")
			}
			log.Error("rootcmd root !=")
			return chain.ErrNotMathcedStateRoot
		}
	} else {
		return ErrGasUsed
	}
	return nil
}

func (bftConsensus *BftConsensus) ReceiveMsg(peer *consensusTypes.PeerInfo, t uint64, buf []byte) {
	switch t {
	case MsgTypeSetUp:
		log.WithField("addr", peer).WithField("code", t).Debug("Receive MsgTypeSetUp msg")
	case MsgTypeChallenge:
		log.WithField("addr", peer).WithField("code", t).Debug("Receive MsgTypeChallenge msg")
	case MsgTypeFail:
		log.WithField("addr", peer).WithField("code", t).Debug("Receive MsgTypeFail msg")
	case MsgTypeCommitment:
		log.WithField("addr", peer).WithField("code", t).Debug("Receive MsgTypeCommitment msg")
	case MsgTypeResponse:
		log.WithField("addr", peer).WithField("code", t).Debug("Receive MsgTypeResponse msg")
	default:
		//return fmt.Errorf("consensus unkonw msg type:%d", msg.Code)
	}
	switch t {
	case MsgTypeSetUp:
		fallthrough
	case MsgTypeChallenge:
		fallthrough
	case MsgTypeFail:
		select {
		case bftConsensus.memberMsgPool <- &MsgWrap{peer, t, buf}:
		default:
		}
	case MsgTypeCommitment:
		fallthrough
	case MsgTypeResponse:
		select {
		case bftConsensus.leaderMsgPool <- &MsgWrap{peer, t, buf}:
		default:
		}

	default:
		//return fmt.Errorf("consensus unkonw msg type:%d", msg.Code)
	}
}

func (bftConsensus *BftConsensus) ChangeTime(interval time.Duration) {
	bftConsensus.WaitTime = interval
}

