package bft

import (
	"encoding/hex"
	"errors"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/chain/store"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/drep-project/drep-chain/app"
	blockMgrService "github.com/drep-project/drep-chain/blockmgr"
	chainService "github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/network/p2p"
	p2pService "github.com/drep-project/drep-chain/network/service"
	accountService "github.com/drep-project/drep-chain/pkgs/accounts/service"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	chainTypes "github.com/drep-project/drep-chain/types"
	"gopkg.in/urfave/cli.v1"
)

var (
	MinerFlag = cli.BoolFlag{
		Name:  "miner",
		Usage: "is miner",
	}
)

type BftConsensusService struct {
	P2pServer        p2pService.P2P                       `service:"p2p"`
	ChainService     chainService.ChainServiceInterface   `service:"chain"`
	BroadCastor      blockMgrService.ISendMessage         `service:"blockmgr"`
	BlockMgrNotifier blockMgrService.IBlockNotify         `service:"blockmgr"`
	BlockGenerator   blockMgrService.IBlockBlockGenerator `service:"blockmgr"`
	DatabaseService  *database.DatabaseService            `service:"database"`
	WalletService    *accountService.AccountService       `service:"accounts"`

	BftConsensus *BftConsensus

	apis   []app.API
	Config *BftConfig

	syncBlockEventSub  event.Subscription
	syncBlockEventChan chan event.SyncBlockEvent
	ConsensusEngine    consensusTypes.IConsensusEngine
	Miner              *secp256k1.PrivateKey
	//During the process of synchronizing blocks, the miner stopped mining
	pauseForSync bool
	start        bool
	peersInfo    map[string]*consensusTypes.PeerInfo
	quit         chan struct{}
}

func (bftConsensusService *BftConsensusService) Name() string {
	return "bft"
}

func (bftConsensusService *BftConsensusService) Api() []app.API {
	return bftConsensusService.apis
}

func (bftConsensusService *BftConsensusService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{MinerFlag}
}

func (bftConsensusService *BftConsensusService) Init(executeContext *app.ExecuteContext) error {
	if executeContext.Cli.GlobalIsSet(MinerFlag.Name) {
		bftConsensusService.Config.StartMiner = executeContext.Cli.GlobalBool(MinerFlag.Name)
	}

	var addPeerFeed event.Feed
	var removePeerFeed event.Feed
	bftConsensusService.BftConsensus = NewBftConsensus(
		bftConsensusService.ChainService,
		bftConsensusService.BlockGenerator,
		bftConsensusService.DatabaseService,
		bftConsensusService.P2pServer,
		bftConsensusService.Config,
		&addPeerFeed,
		&removePeerFeed,
	)

	bftConsensusService.ChainService.AddBlockValidator(&BlockMultiSigValidator{bftConsensusService.BftConsensus.GetProducers, bftConsensusService.ChainService.GetBlockByHash, bftConsensusService.Config.ProducerNum})
	bftConsensusService.ChainService.AddGenesisProcess(NewMinerGenesisProcessor())
	if !bftConsensusService.Config.StartMiner {
		return nil
	} else {
		if bftConsensusService.WalletService.Wallet == nil {
			return ErrWalletNotOpen
		}
	}

	//consult privkey in wallet
	accountNode, err := bftConsensusService.WalletService.Wallet.GetAccountByPubkey(bftConsensusService.Config.MyPk)
	if err != nil {
		log.WithField("init err", err).WithField("addr", crypto.PubkeyToAddress(bftConsensusService.Config.MyPk).String()).Error("privkey of MyPk in Config is not in local wallet")
		return err
	}
	bftConsensusService.Miner = accountNode.PrivateKey
	bftConsensusService.P2pServer.AddProtocols([]p2p.Protocol{
		p2p.Protocol{
			Name:   "bftConsensusService",
			Length: NumberOfMsg,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				defer func() {
					log.WithField("IP", peer.Node().IP().String()).
						WithField("PublicKey", hex.EncodeToString(peer.Node().Pubkey().Serialize())).
						Debug("bft protocol err, disconnect peer")
					peer.Disconnect(p2p.DiscQuitting)
				}()
				producers, err := bftConsensusService.GetProducers(bftConsensusService.ChainService.BestChain().Tip().Height, bftConsensusService.Config.ProducerNum*2)
				if err != nil {
					log.WithField("err", err).Info("fail to get producers")
					return err
				}

				ipChecked := false
				for _, producer := range producers {
					if producer.Node.IP().String() == peer.Node().IP().String() {
						ipChecked = true
						break
					}
				}
				if !ipChecked {
					log.WithField("IP", peer.Node().IP().String()).
						WithField("PublicKey", hex.EncodeToString(peer.Node().Pubkey().Serialize())).
						Debug("Receive remove peer")
					for _, producer := range producers {
						log.WithField("IP", producer.Node.IP().String()).
							WithField("PublicKey", hex.EncodeToString(producer.Node.Pubkey().Serialize())).
							Debug("Exit Candidate peer")
					}
					return ErrBpNotInList
				}
				pi := consensusTypes.NewPeerInfo(peer, rw)
				//send verify message
				randomBytes := [32]byte{}
				rand.Read(randomBytes[:])
				err = bftConsensusService.P2pServer.Send(rw, MsgTypeValidateReq, randomBytes)
				if err != nil {
					return err
				}
				//del peer event
				ch := make(chan *p2p.PeerEvent)
				sub := bftConsensusService.P2pServer.SubscribeEvents(ch)
				//control producer validator by timer
				tm := time.NewTimer(time.Second * 10)
				defer func() {
					removePeerFeed.Send(pi)
					sub.Unsubscribe()
				}()
				for {
					select {
					case e := <-ch:
						if e.Type == p2p.PeerEventTypeDrop {
							return errors.New(e.Error)
						}
					case <-tm.C:
						return errors.New("timeout: wait validata message")
					default:
						msg, err := rw.ReadMsg()
						if err != nil {
							log.WithField("Reason", err).WithField("Ip", pi.IP()).Error("consensus receive msg")
							return err
						}
						buf, err := ioutil.ReadAll(msg.Payload)
						if err != nil {
							return err
						}

						switch msg.Code {
						//	case bftConsensusService.P2pServer
						case MsgTypeValidateReq:
							if err != nil {
								return err
							}
							sig, err := accountNode.PrivateKey.Sign(buf)
							if err != nil {
								return err
							}
							err = bftConsensusService.P2pServer.Send(rw, MsgTypeValidateRes, sig)
							if err != nil {
								return err
							}
						case MsgTypeValidateRes:
							sig := &secp256k1.Signature{}
							err := binary.Unmarshal(buf, sig)
							if err != nil {
								return err
							}

							for _, producer := range producers {
								if sig.Verify(randomBytes[:], producer.Pubkey) {
									addPeerFeed.Send(pi)
									tm.Stop()
								}
							}
							continue
						default:
							bftConsensusService.BftConsensus.ReceiveMsg(pi, msg.Code, buf)
						}
					}
				}
			},
		},
	})
	bftConsensusService.syncBlockEventChan = make(chan event.SyncBlockEvent)
	bftConsensusService.syncBlockEventSub = bftConsensusService.BlockMgrNotifier.SubscribeSyncBlockEvent(bftConsensusService.syncBlockEventChan)
	bftConsensusService.quit = make(chan struct{})
	bftConsensusService.apis = []app.API{
		app.API{
			Namespace: "consensus",
			Version:   "1.0",
			Service: &ConsensusApi{
				consensusService: bftConsensusService,
			},
			Public: true,
		},
	}

	go bftConsensusService.handlerEvent()
	return nil
}

func (bftConsensusService *BftConsensusService) handlerEvent() {
	for {
		select {
		case e := <-bftConsensusService.syncBlockEventChan:
			if e.EventType == event.StartSyncBlock {
				bftConsensusService.pauseForSync = true
				log.Info("Start Sync Blcok")
			} else {
				bftConsensusService.pauseForSync = false
				log.Info("Stop Sync Blcok")
			}
		case <-bftConsensusService.quit:
			return
		}
	}
}

func (bftConsensusService *BftConsensusService) Start(executeContext *app.ExecuteContext) error {
	if !bftConsensusService.Config.StartMiner {
		return nil
	}
	bftConsensusService.start = true
	go func() {
		select {
		case <-bftConsensusService.quit:
			return
		default:
			for {
				if bftConsensusService.pauseForSync {
					time.Sleep(time.Millisecond * 500)
					continue
				}
				log.WithField("Height", bftConsensusService.ChainService.BestChain().Height()).Trace("node start")
				block, err := bftConsensusService.BftConsensus.Run(bftConsensusService.Miner)
				if err != nil {
					log.WithField("Reason", err.Error()).Debug("Producer Block Fail")
				} else {
					_, _, err := bftConsensusService.ChainService.ProcessBlock(block)
					if err == nil {
						bftConsensusService.BroadCastor.BroadcastBlock(chainTypes.MsgTypeBlock, block, true)
						log.WithField("Height", block.Header.Height).WithField("txs:", block.Data.TxCount).Info("Process block successfully and broad case block message")
					} else {
						log.WithField("Height", block.Header.Height).WithField("txs:", block.Data.TxCount).WithField("err", err).Info("Process Block fail")
					}
				}
				nextBlockTime, waitSpan := bftConsensusService.getWaitTime()
				log.WithField("nextBlockTime", nextBlockTime).WithField("waitSpan", waitSpan).Debug("Sleep")
				time.Sleep(waitSpan)
			}
		}
	}()

	return nil
}

func (bftConsensusService *BftConsensusService) Stop(executeContext *app.ExecuteContext) error {
	if bftConsensusService.Config == nil || !bftConsensusService.Config.StartMiner {
		return nil
	}

	if bftConsensusService.quit != nil {
		close(bftConsensusService.quit)
	}

	if bftConsensusService.syncBlockEventSub != nil {
		bftConsensusService.syncBlockEventSub.Unsubscribe()
	}

	return nil
}

func (bftConsensusService *BftConsensusService) getWaitTime() (time.Time, time.Duration) {
	lastBlockTime := time.Unix(int64(bftConsensusService.ChainService.BestChain().Tip().TimeStamp), 0)
	targetTime := lastBlockTime.Add(time.Duration(int64(time.Second) * bftConsensusService.Config.BlockInterval))
	now := time.Now()
	if targetTime.Before(now) {
		interval := now.Sub(lastBlockTime)
		nextBlockInterval := int64(interval/(time.Second * time.Duration(bftConsensusService.Config.BlockInterval))) + 1
		nextBlockTime := lastBlockTime.Add(time.Second * time.Duration(nextBlockInterval *  bftConsensusService.Config.BlockInterval))
		return nextBlockTime, nextBlockTime.Sub(now)
	} else {
		return targetTime, targetTime.Sub(now)
	}
}

func (bftConsensusService *BftConsensusService) GetProducers(height uint64, topN int) ([]*Producer, error) {
	block, err := bftConsensusService.ChainService.GetBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	trie, err := store.TrieStoreFromStore(bftConsensusService.DatabaseService.LevelDb(), block.Header.StateRoot)
	if err != nil {
		return nil, err
	}
	return GetCandidates(trie, topN), nil
}

func (bftConsensusService *BftConsensusService) DefaultConfig() *BftConfig {
	return &BftConfig{
		BlockInterval:  10,
		ProducerNum:    7,
		ChangeInterval: 100,
	}
}
