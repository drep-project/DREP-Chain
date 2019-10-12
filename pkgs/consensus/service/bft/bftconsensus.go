package bft

import (
	"bytes"
	"github.com/drep-project/drep-chain/chain/store"
	"math"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/blockmgr"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/params"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"github.com/drep-project/drep-chain/types"
)

const (
	waitTime = 10 * time.Second
)

type BftConsensus struct {
	CoinBase  crypto.CommonAddress
	PrivKey   *secp256k1.PrivateKey
	curMiner  int
	minMiners int

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
	Producers      ProducerSet
}

func NewBftConsensus(
	chainService chain.ChainServiceInterface,
	blockGenerator blockmgr.IBlockBlockGenerator,
	dbService *database.DatabaseService,
	producer ProducerSet,
	sener Sender,
	addPeer, removePeer *event.Feed) *BftConsensus {

	addPeerChan := make(chan *consensusTypes.PeerInfo)
	removePeerChan := make(chan *consensusTypes.PeerInfo)
	addPeer.Subscribe(addPeerChan)
	removePeer.Subscribe(removePeerChan)
	return &BftConsensus{
		BlockGenerator: blockGenerator,
		ChainService:   chainService,
		DbService:      dbService,
		minMiners:      int(math.Ceil(float64(len(producer)) * 2 / 3)),
		Producers:      producer,
		sender:         sener,
		onLinePeer:     map[string]consensusTypes.IPeerInfo{},
		WaitTime:       waitTime,
		memberMsgPool:  make(chan *MsgWrap, 1000),
		leaderMsgPool:  make(chan *MsgWrap, 1000),
		addPeerChan:    addPeerChan,
		removePeerChan: removePeerChan,
	}
}

func (bftConsensus *BftConsensus) Run(privKey *secp256k1.PrivateKey) (*types.Block, error) {
	bftConsensus.CoinBase = crypto.PubkeyToAddress(privKey.PubKey())
	bftConsensus.PrivKey = privKey
	go bftConsensus.processPeers()
	miners := bftConsensus.collectMemberStatus()
	if len(miners) > 1 {
		isM, isL := bftConsensus.moveToNextMiner(miners)
		if isL {
			return bftConsensus.runAsLeader(miners)
		} else if isM {
			return bftConsensus.runAsMember(miners)
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

	if curMiner.IsMe {
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
			pi           consensusTypes.IPeerInfo
		)

		isMe := bftConsensus.PrivKey.PubKey().IsEqual(produce.Pubkey)
		if isMe {
			IsOnline = true
		} else {
			//todo  peer获取到的IP地址和配置的ip地址是否相等（nat后是否相等,从tcp原理来看是相等的）
			bftConsensus.peerLock.RLock()
			if pi, ok = bftConsensus.onLinePeer[produce.IP]; ok {
				IsOnline = true
			}
			bftConsensus.peerLock.RUnlock()
		}

		produceInfos = append(produceInfos, &MemberInfo{
			Producer: &Producer{Pubkey: produce.Pubkey, IP: produce.IP},
			Peer:     pi,
			IsMe:     isMe,
			IsOnline: IsOnline,
		})
	}
	return produceInfos
}

func (bftConsensus *BftConsensus) runAsMember(miners []*MemberInfo) (block *types.Block, err error) {
	member := NewMember(bftConsensus.PrivKey, bftConsensus.sender, bftConsensus.WaitTime, miners, bftConsensus.minMiners, bftConsensus.ChainService.BestChain().Height(), bftConsensus.memberMsgPool)
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
func (bftConsensus *BftConsensus) runAsLeader(miners []*MemberInfo) (block *types.Block, err error) {
	leader := NewLeader(
		bftConsensus.PrivKey,
		bftConsensus.sender,
		bftConsensus.WaitTime,
		miners,
		bftConsensus.minMiners,
		bftConsensus.ChainService.BestChain().Height(),
		bftConsensus.leaderMsgPool)
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
	err = AccumulateRewards(trieStore, multiSig, bftConsensus.Producers, gasFee, block.Header.Height)
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
	multiSigValidator := BlockMultiSigValidator{bftConsensus.Producers}
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
	log.WithField("addr", peer).WithField("code", t).Debug("Receive setup msg")
	switch t {
	case MsgTypeSetUp:
		fallthrough
	case MsgTypeChallenge:
		fallthrough
	case MsgTypeFail:
		bftConsensus.memberMsgPool <- &MsgWrap{peer, t, buf}
	case MsgTypeCommitment:
		fallthrough
	case MsgTypeResponse:
		bftConsensus.leaderMsgPool <- &MsgWrap{peer, t, buf}
	default:
		//return fmt.Errorf("consensus unkonw msg type:%d", msg.Code)
	}
}

func (bftConsensus *BftConsensus) ChangeTime(interval time.Duration) {
	bftConsensus.WaitTime = interval
}

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func AccumulateRewards(trieStore store.StoreInterface, sig *MultiSignature, Producers ProducerSet, totalGasBalance *big.Int, height uint64) error {
	reward := new(big.Int).SetUint64(uint64(params.Rewards))
	r := new(big.Int)
	r = r.Div(reward, new(big.Int).SetInt64(2))
	r.Add(r, totalGasBalance)
	leaderAddr := Producers[sig.Leader].Address()
	err := trieStore.AddBalance(&leaderAddr, height, r)
	if err != nil {
		return err
	}

	num := sig.Num() - 1
	for index, isCommit := range sig.Bitmap {
		if isCommit == 1 {
			addr := Producers[index].Address()
			if addr != leaderAddr {
				r.Div(reward, new(big.Int).SetInt64(int64(num*2)))
				err = trieStore.AddBalance(&addr, height, r)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}