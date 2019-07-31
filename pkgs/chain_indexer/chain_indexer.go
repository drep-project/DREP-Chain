package chain_indexer

import (
	"fmt"
	"context"
	"time"
	"sync"

	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
	"github.com/drep-project/drep-chain/database"
)

// ChainIndexerChain interface is used for connecting the indexer to a blockchain
type ChainIndexerChain interface {
	// GetCurrentHeader retrieves the latest locally known header.
	GetCurrentHeader() *types.BlockHeader

	// NewBlockFeed return feed to subscribe new head header notifications.
	NewBlockFeed() *event.Feed
}

type ChainIndexerServiceInterface interface {
	app.Service
}

var _ ChainIndexerServiceInterface = &ChainIndexerService{}

type ChainIndexerService struct {
	DatabaseService *database.DatabaseService		`service:"database"`
	ChainService	chain.ChainServiceInterface		`service:"chain"`
	chainId types.ChainIdType

	Config *ChainIndexerConfig

	storedSections uint64 // Number of sections successfully indexed into the database
	knownSections  uint64 // Number of sections known to be complete (block wise)
	cascadedHead   uint64 // Block number of the last completed section cascaded to subindexers

	checkpointSections uint64      // Number of sections covered by the checkpoint
	checkpointHead     crypto.Hash // Section head belonging to the checkpoint

	active    uint32          // Flag whether the event loop was started
	update    chan struct{}   // Notification channel that headers should be processed
	quit      chan chan error // Quit channel to tear down running goroutines
	ctx       context.Context
	ctxCancel func()

	lock sync.RWMutex
}

func (chainIndexer *ChainIndexerService) Name() string {
	return MODULENAME
}

func (chainIndexer *ChainIndexerService) Api() []app.API {
	return []app.API{}
}

func (chainIndexer *ChainIndexerService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{EnableChainIndexerFlag}
}

func (chainIndexer *ChainIndexerService) Init(executeContext *app.ExecuteContext) error {
	// check service dependencies
	if chainIndexer.DatabaseService == nil {
		return fmt.Errorf("batabaseService not init")
	}
	if chainIndexer.ChainService == nil {
		return fmt.Errorf("chainService not init")
	}

	// initialize module config
	chainIndexer.Config = DefaultConfig
	err := executeContext.UnmashalConfig(chainIndexer.Name(), chainIndexer.Config)
	if err != nil {
		return err
	}
	if executeContext.Cli.GlobalIsSet(EnableChainIndexerFlag.Name) {
		chainIndexer.Config.Enable = executeContext.Cli.GlobalBool(EnableChainIndexerFlag.Name)
	}
	if !chainIndexer.Config.Enable {
		return nil
	}

	// initialize other fields in chainIndexer
	chainIndexer.update = make(chan struct{}, 1)
	chainIndexer.quit = make(chan chan error)
	chainIndexer.ctx, chainIndexer.ctxCancel = context.WithCancel(context.Background())
	chainIndexer.storedSections = chainIndexer.GetStoredSections()

	go chainIndexer.updateLoop()

	return nil
}

func (chainIndexer *ChainIndexerService) Start(executeContext *app.ExecuteContext) error {

	return nil
}

func (chainIndexer *ChainIndexerService) Stop(executeContext *app.ExecuteContext) error {

	return nil
}

// updateLoop is the main event loop of the indexer which pushes chain segments
// down into the processing backend.
func (chainIndexer *ChainIndexerService) updateLoop() {
	var (
		updating bool
		updated  time.Time
	)

	for {
		select {
		case errc := <-chainIndexer.quit:
			// Chain indexer terminating, report no failure and abort
			errc <- nil
			return

		case <-chainIndexer.update:
			// Section headers completed (or rolled back), update the index
			chainIndexer.lock.Lock()
			if chainIndexer.knownSections > chainIndexer.storedSections {
				// Periodically print an upgrade log message to the user
				if time.Since(updated) > 8*time.Second {
					if chainIndexer.knownSections > chainIndexer.storedSections+1 {
						updating = true
						log.WithField("percentage", chainIndexer.storedSections * 100 / chainIndexer.knownSections).Info("Upgrading chain index")
					}
					updated = time.Now()
				}
				// Cache the current section count and head to allow unlocking the mutex
				chainIndexer.verifyLastHead()
				section := chainIndexer.storedSections
				var oldHead crypto.Hash
				if section > 0 {
					oldHead = chainIndexer.SectionHead(section - 1)
				}
				// Process the newly defined section in the background
				chainIndexer.lock.Unlock()
				newHead, err := chainIndexer.processSection(section, oldHead)
				if err != nil {
					select {
					case <-chainIndexer.ctx.Done():
						<-chainIndexer.quit <- nil
						return
					default:
					}
					log.WithField("error", err).Error("Section processing failed")
				}
				chainIndexer.lock.Lock()

				// If processing succeeded and no reorgs occurred, mark the section completed
				if err == nil && (section == 0 || oldHead == chainIndexer.SectionHead(section-1)) {
					chainIndexer.setSectionHead(section, newHead)
					chainIndexer.setValidSections(section + 1)
					if chainIndexer.storedSections == chainIndexer.knownSections && updating {
						updating = false
						log.Info("Finished upgrading chain index")
					}

				} else {
					// If processing failed, don't retry until further notification
					log.Debug("Chain index processing failed", "section", section, "err", err)
					chainIndexer.verifyLastHead()
					chainIndexer.knownSections = chainIndexer.storedSections
				}
			}
			// If there are still further sections to process, reschedule
			if chainIndexer.knownSections > chainIndexer.storedSections {
				time.AfterFunc(chainIndexer.Config.Throttling, func() {
					select {
					case chainIndexer.update <- struct{}{}:
					default:
					}
				})
			}
			chainIndexer.lock.Unlock()
		}
	}
}

// verifyLastHead compares last stored section head with the corresponding block hash in the
// actual canonical chain and rolls back reorged sections if necessary to ensure that stored
// sections are all valid
func (chainIndexer *ChainIndexerService) verifyLastHead() {
	for chainIndexer.storedSections > 0 && chainIndexer.storedSections > chainIndexer.checkpointSections {

		hash := crypto.Hash{}
		blockHeader, err := chainIndexer.ChainService.GetBlockHeaderByHeight(chainIndexer.storedSections * chainIndexer.Config.SectionSize - 1)
		if err == nil {
			hash = *blockHeader.Hash()
		}

		if chainIndexer.SectionHead(chainIndexer.storedSections-1) == hash {
			return
		}
		chainIndexer.setValidSections(chainIndexer.storedSections - 1)
	}
}

// SectionHead 从数据库中获取已处理section的最后一个块哈希。
func (chainIndexer *ChainIndexerService) SectionHead(section uint64) crypto.Hash {
	//var data [8]byte
	//binary.BigEndian.PutUint64(data[:], section)
	//
	//hash, _ := chainIndexer.indexDb.Get(append([]byte("shead"), data[:]...))
	//if len(hash) == len(crypto.Hash{}) {
	//	return crypto.BytesToHash(hash)
	//}
	return crypto.Hash{}
}

//setSectionHead 将已处理section的最后一个块哈希写入数据库。
func (chainIndexer *ChainIndexerService) setSectionHead(section uint64, hash crypto.Hash) {
	//var data [8]byte
	//binary.BigEndian.PutUint64(data[:], section)
	//
	//chainIndexer.indexDb.Put(append([]byte("shead"), data[:]...), hash.Bytes())

	return
}

// processSection processes an entire section by calling backend functions while
// ensuring the continuity of the passed headers. Since the chain mutex is not
// held while processing, the continuity can be broken by a long reorg, in which
// case the function returns with an error.
func (chainIndexer *ChainIndexerService) processSection(section uint64, lastHead crypto.Hash) (crypto.Hash, error) {
	return crypto.Hash{}, nil
}

// setValidSections writes the number of valid sections to the index database
func (chainIndexer *ChainIndexerService) setValidSections(sections uint64) {
	return
}