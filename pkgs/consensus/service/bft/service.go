package bft

import (
	"errors"
	"fmt"
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

	bftConsensusService.ChainService.AddBlockValidator(&BlockMultiSigValidator{bftConsensusService.BftConsensus.GetProducers, bftConsensusService.ChainService.GetBlockByHash,bftConsensusService.Config.ProducerNum})
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
			Length: NumberOfMsg + 2,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				MsgTypeValidateReq := uint64(NumberOfMsg)
				MsgTypeValidateRes := uint64(NumberOfMsg + 1)
				pi := consensusTypes.NewPeerInfo(peer, rw)
				//send verify message
				randomBytes := [32]byte{}
				rand.Read(randomBytes[:])
				err := bftConsensusService.P2pServer.Send(rw, MsgTypeValidateReq, randomBytes)
				if err != nil {
					return err
				}
				fmt.Println(pi.IP())
				//del peer event
				ch := make(chan *p2p.PeerEvent)
				//sub := bftConsensusService.P2pServer.SubscribeEvents(ch)
				//control producer validator by timer
				tm := time.NewTimer(time.Second * 10)
				defer func() {
					peer.Disconnect(p2p.DiscQuitting)
					removePeerFeed.Send(pi)
					//sub.Unsubscribe()
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
							producers, err := bftConsensusService.GetProducers(bftConsensusService.ChainService.BestChain().Tip().Height, bftConsensusService.Config.ProducerNum*2)
							if err != nil {
								log.WithField("err", err).Info("get producers")
								//return err
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

				log.WithField("peer.ip", peer.IP()).Info("peer not producer")
				//非骨干节点，不启动共识相关处理
				return nil
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
	// max_delay_time +(min_block_interval)*windows = expected_block_interval*windows
	// 6h + 5s*windows = 10s*windows
	// windows = 4320

	lastBlockTime := time.Unix(int64(bftConsensusService.ChainService.BestChain().Tip().TimeStamp), 0)
	targetTime := lastBlockTime.Add(time.Duration(int(time.Second)*bftConsensusService.Config.BlockInterval))
	now := time.Now()
	if targetTime.Before(now) {
		return now.Add(time.Millisecond * 500), time.Millisecond * 500
	} else {
		return targetTime, targetTime.Sub(now)
	}
	/*
		     window := int64(4320)
		     endBlock := bftConsensusService.DatabaseService.GetHighestBlock().Header
		     if endBlock.Height < window {
				 lastBlockTime := time.Unix(bftConsensusService.DatabaseService.GetHighestBlock().Header.Timestamp, 0)
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
				 startBlock :=bftConsensusService.DatabaseService.GetBlock(startHeight).Header

				 xx := window * 10 -(time.Unix(startBlock.Timestamp,0).Sub(time.Unix(endBlock.Timestamp,0))).Seconds()

				 span := time.Unix(startBlock.Timestamp,0).Sub(time.Unix(endBlock.Timestamp,0))  //window time
				 avgSpan := span.Nanoseconds()/window
				 return time.Duration(avgSpan) * time.Nanosecond
			 }
	*/
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
		BlockInterval: int(time.Second * 5),
		ProducerNum: 7,
		ChangeInterval:1000,
	}
}
