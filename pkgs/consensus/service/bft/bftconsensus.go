package bft

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/drep-project/DREP-Chain/blockmgr"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common/event"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/database"
	p2pService "github.com/drep-project/DREP-Chain/network/service"
	consensusTypes "github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	"github.com/drep-project/DREP-Chain/types"
	drepbinary "github.com/drep-project/binary"
)

const (
	waitTime = 10 * time.Second
	round1   = 1
	round2   = 2
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
	onLinePeer map[string]consensusTypes.IPeerInfo //key: enode.ID，value ,peerInfo
	WaitTime   time.Duration

	memberMsgPool chan *MsgWrap
	leaderMsgPool chan *MsgWrap

	addPeerChan    chan *consensusTypes.PeerInfo
	removePeerChan chan *consensusTypes.PeerInfo

	producer []types.Producer
	quit     chan struct{}

	chBestHeight chan uint64
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

	value := make([]byte, 0)
	buffer := bytes.NewBuffer(value)
	binary.Write(buffer, binary.BigEndian, config.ChangeInterval)
	dbService.LevelDb().Put([]byte(store.ChangeInterval), buffer.Bytes())

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
		memberMsgPool:  make(chan *MsgWrap, 1000),
		leaderMsgPool:  make(chan *MsgWrap, 1000),
		quit:           make(chan struct{}),
		chBestHeight:   make(chan uint64, 0),
	}
}

func (bftConsensus *BftConsensus) GetProducers(height uint64, topN int) ([]types.Producer, error) {
	//newEpoch := height % uint64(bftConsensus.config.ChangeInterval)
	//if bftConsensus.producer == nil || newEpoch == 0 {
	//height = height - newEpoch

	producers, err := bftConsensus.loadProducers(height, topN)
	bftConsensus.producer = producers
	return producers, err
	//} else {
	//	return bftConsensus.producer, nil
	//}
}

func (bftConsensus *BftConsensus) loadProducers(height uint64, topN int) ([]types.Producer, error) {
	block, err := bftConsensus.ChainService.GetBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	trie, err := store.TrieStoreFromStore(bftConsensus.DbService.LevelDb(), block.Header.StateRoot)
	if err != nil {
		return nil, err
	}

	return GetCandidates(trie, topN), nil
}

func (bftConsensus *BftConsensus) clearMsgPool() {
	//for {
	//	select {
	//	case <-bftConsensus.memberMsgPool:
	//		fmt.Println("clean peer leader msg......")
	//	case <-bftConsensus.leaderMsgPool:
	//		fmt.Println("clean peer member msg .....")
	//	default:
	//		return
	//	}
	//}
}

func (bftConsensus *BftConsensus) Run(privKey *secp256k1.PrivateKey) (*types.Block, error) {
	bftConsensus.CoinBase = crypto.PubkeyToAddress(privKey.PubKey())
	bftConsensus.PrivKey = privKey

	producers, err := bftConsensus.GetProducers(bftConsensus.ChainService.BestChain().Height(), bftConsensus.config.ProducerNum)
	if err != nil {
		log.Trace("bft consensus run get producers err:", err)
		return nil, err
	}
	found := false

	log.Trace(" bftConsensus.config.ProducerNum:", bftConsensus.config.ProducerNum, bftConsensus.ChainService.BestChain().Height())
	for _, p := range producers {
		log.WithField("node", p.Node.String()).Trace("get producers")
	}

	for _, p := range producers {
		if bytes.Equal(p.Pubkey.Serialize(), bftConsensus.config.MyPk.Serialize()) {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrNotMyTurn
	}

	minMiners := int(bftConsensus.config.ProducerNum * 2 / 3)
	if bftConsensus.config.ProducerNum*2%3 != 0 {
		minMiners++
	}
	miners := bftConsensus.collectMemberStatus(producers)
	//print miners status
	str := "-----------------------------------\n"
	for _, m := range miners {
		str += "|	" + m.Producer.Node.IP().String() + "|	" + strconv.FormatBool(m.IsOnline) + "	|\n"
	}
	fmt.Println(str)

	if len(miners) > 1 {
		isM, isL, err := bftConsensus.moveToNextMiner(miners)
		if err != nil {
			return nil, err
		}
		log.WithField("isL", isL).Trace("BftConsensus run")
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
			bftConsensus.onLinePeer[addPeer.ID()] = addPeer
			bftConsensus.peerLock.Unlock()
			log.WithField("bft new peer:", addPeer.IP()).Info("process peer")
		case removePeer := <-bftConsensus.removePeerChan:
			bftConsensus.peerLock.Lock()
			delete(bftConsensus.onLinePeer, removePeer.ID())
			bftConsensus.peerLock.Unlock()
			log.Info("bft remove peer:", removePeer.IP())

		case <-bftConsensus.quit:
			return
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

	//如果新来的块时间与系统的时间差距较小，可能是此轮出块已经结束，当前最新的块是由其它节点达成的共识；
	// 此块同步到本地后，恰好本地新的出块循环被触发。出现此种现象的原因是：定时器或者每个节点的时间并不是完全同步的
	if int64(bftConsensus.ChainService.BestChain().Tip().TimeStamp) >= time.Now().Unix() ||
		time.Now().Unix()-int64(bftConsensus.ChainService.BestChain().Tip().TimeStamp) < int64(bftConsensus.config.BlockInterval/2) {
		log.WithField("now", time.Now().Unix()).WithField("bestBlock ts",
			bftConsensus.ChainService.BestChain().Tip().TimeStamp).Trace("moveToNextMiner ts err")
		return false, false, fmt.Errorf("new block time err")
	}

	if curMiner.IsMe {
		return false, true, nil
	} else {
		return true, false, nil
	}
}

func (bftConsensus *BftConsensus) collectMemberStatus(producers []types.Producer) []*MemberInfo {
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
			bftConsensus.peerLock.RLock()
			if pi, ok = bftConsensus.onLinePeer[produce.Node.ID().String()]; ok {
				IsOnline = true
			}
			bftConsensus.peerLock.RUnlock()
		}

		produceInfos = append(produceInfos, &MemberInfo{
			Producer: &types.Producer{Pubkey: produce.Pubkey, Node: produce.Node},
			Peer:     pi,
			IsMe:     isMe,
			IsOnline: IsOnline,
		})
	}
	return produceInfos
}

func (bftConsensus *BftConsensus) bestHeight() {
	tm := time.NewTicker(time.Millisecond * 500)
	defer tm.Stop()
	for {
		select {
		case <-bftConsensus.quit:
			return

		case <-tm.C:
			bftConsensus.chBestHeight <- bftConsensus.ChainService.BestChain().Height()
		}
	}
}

func (bftConsensus *BftConsensus) runAsMember(miners []*MemberInfo, minMiners int) (block *types.Block, err error) {
	member := NewMember(bftConsensus.PrivKey, bftConsensus.sender, bftConsensus.WaitTime, miners, minMiners,
		bftConsensus.ChainService.BestChain().Height(), bftConsensus.memberMsgPool)
	log.Trace("node member is going to process consensus for round 1")
	member.convertor = func(msg []byte) (IConsenMsg, error) {
		block, err = types.BlockFromMessage(msg)
		if err != nil {
			return nil, err
		}

		if block == nil {
			return nil, fmt.Errorf("unmarshal msg err")
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
	_, err = member.ProcessConsensus(round1, bftConsensus.chBestHeight)
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
		multiSigBytes, err := drepbinary.Marshal(multiSig)
		if err != nil {
			log.Error("fail to marshal MultiSig")
			return err
		}
		log.WithField("bitmap", multiSig.Bitmap).Info("member receive participant bitmap")
		block.Proof = types.Proof{Type: consensusTypes.Pbft, Evidence: multiSigBytes}
		return bftConsensus.verifyBlockContent(block)
	}
	_, err = member.ProcessConsensus(round2, bftConsensus.chBestHeight)
	if err != nil {
		return nil, err
	}
	member.Reset()
	log.Trace("node member finishes consensus for round 2")
	return block, nil
}

//1 The leader makes blocks, signs them and broadcasts them to other producers,
//2 Other producers will sign their own digital signatures after receiving them. The signed block is then returned to the leader
//3 After the leader collects all the signatures or returns more than two-thirds of the number of producers, he or she shall verify the signatures
//4 After the leader validates the signature, the block is broadcast to all peers
func (bftConsensus *BftConsensus) runAsLeader(producers types.ProducerSet, miners []*MemberInfo, minMiners int) (block *types.Block, err error) {

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
		log.WithField("err", err).Trace("reun As Leader")
		return nil, err
	}
	var gasFee *big.Int
	block, gasFee, err = bftConsensus.BlockGenerator.GenerateTemplate(trieStore, bftConsensus.CoinBase,
		int(bftConsensus.config.BlockInterval))
	if err != nil {
		log.WithField("msg", err).Error("generate block fail")
		return nil, err
	}

	log.WithField("Block", block).Trace("node leader is preparing process consensus for round 1")
	err, sig, bitmap := leader.ProcessConsensus(block, round1, bftConsensus.chBestHeight)
	if err != nil {
		var str = err.Error()
		log.WithField("msg", str).WithField("round", round1).Error("Error occurs")
		return nil, err
	}

	leader.Reset()
	multiSig := newMultiSignature(*sig, bftConsensus.curMiner, bitmap)
	multiSigBytes, err := drepbinary.Marshal(multiSig)
	if err != nil {
		log.Error("fial to marshal MultiSig")
		return nil, err
	}
	log.WithField("bitmap", multiSig.Bitmap).Info("participant bitmap")
	//Determine reward points
	block.Proof = types.Proof{Type: consensusTypes.Pbft, Evidence: multiSigBytes}
	calculator := NewRewardCalculator(trieStore, multiSig, producers, gasFee, block.Header.Height)
	err = calculator.AccumulateRewards(block.Header.Height)
	if err != nil {
		log.WithField("err", err).WithField("height", block.Header.Height).Info("accumulate rewards")
		return nil, err
	}

	block.Header.StateRoot = trieStore.GetStateRoot()
	rwMsg := &CompletedBlockMessage{*multiSig, block.Header.StateRoot}

	log.Trace("node leader is going to process consensus for round 2")
	err, _, _ = leader.ProcessConsensus(rwMsg, round2, bftConsensus.chBestHeight)
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
	validator := bftConsensus.ChainService.BlockValidator().SelectByType(reflect.TypeOf(chain.ChainBlockValidator{}))
	err = validator.VerifyHeader(block.Header, preBlockHash)
	if err != nil {
		return err
	}
	err = validator.VerifyBody(block)
	if err != nil {
		return err
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
	validators := bftConsensus.ChainService.BlockValidator()
	for _, validator := range validators {
		err = validator.ExecuteBlock(context)
		if err != nil {
			return err
		}
	}

	multiSig := &MultiSignature{}
	err = drepbinary.Unmarshal(block.Proof.Evidence, multiSig)
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
		log.WithField("addr", peer.IP()).WithField("code", t).WithField("size", len(buf)).Debug("Receive MsgTypeSetUp msg")
	case MsgTypeChallenge:
		log.WithField("addr", peer.IP()).WithField("code", t).WithField("size", len(buf)).Debug("Receive MsgTypeChallenge msg")
	//case MsgTypeFail:
	//	f := Fail{}
	//	drepbinary.Unmarshal(buf, &f)
	//	log.WithField("addr", peer.IP()).WithField("code", t).WithField("buf", f).Debug("Receive MsgTypeFail msg")
	case MsgTypeCommitment:
		log.WithField("addr", peer.IP()).WithField("code", t).Debug("Receive MsgTypeCommitment msg")
	case MsgTypeResponse:
		log.WithField("addr", peer.IP()).WithField("code", t).Debug("Receive MsgTypeResponse msg")
	default:
		//return fmt.Errorf("consensus unkonw msg type:%d", msg.Code)
	}
	switch t {
	case MsgTypeSetUp:
		fallthrough
	case MsgTypeChallenge:
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

func (bftConsensus *BftConsensus) Close() {
	close(bftConsensus.quit)
}

func (bftConsensus *BftConsensus) prepareForMining(p2p p2pService.P2P) {
	if bftConsensus.config.MyPk == nil {
		log.Trace("not set pubkey ,No need to start threads")
		return
	}
	for {
		timer := time.NewTicker(time.Second * 3)
		defer timer.Stop()

		select {
		case <-timer.C:
			//Get as many candidate nodes as possible, establish connection with other candidate nodes in advance, and prepare for the next block
			producers, err := bftConsensus.loadProducers(bftConsensus.ChainService.BestChain().Tip().Height, bftConsensus.config.ProducerNum*3/2)
			if err != nil {
				log.WithField("err", err).Info("PrepareForMiner get producer err")
			}

			tempProduces := make([]types.Producer, len(producers))
			copy(tempProduces, producers)
			//I'm in the running
			found := false
			for index, p := range tempProduces {
				if bytes.Equal(p.Pubkey.Serialize(), bftConsensus.config.MyPk.Serialize()) {
					found = true
					tempProduces = append(tempProduces[:index], tempProduces[index+1:]...)
					break
				}
			}

			if found {
				for _, p := range tempProduces {
					if _, ok := bftConsensus.onLinePeer[p.Node.ID().String()]; !ok {
						p2p.RemovePeer(p.Node.String())
						p2p.AddPeer(p.Node.String())
						log.Trace("prepare for mining,add peer:", p.Node.String())
					}
				}
			}

		case <-bftConsensus.quit:
			return
		}
	}
}
