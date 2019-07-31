package chain_indexer

import (
	"fmt"
	"context"
	"time"
	"sync"

	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/common/bitutil"
	"github.com/drep-project/drep-chain/common/bloombits"
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

	// bloomIndexer
	size    uint64		// section size to generate bloombits for
	gen     *bloombits.Generator // generator to rotate the bloom bits crating the bloom index
	section uint64               // Section is the section number being processed currently
	head    crypto.Hash          // Head is the hash of the last header processed
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
	chainIndexer.storedSections = chainIndexer.getStoredSections()

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
					oldHead = chainIndexer.getSectionHead(section - 1)
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
				if err == nil && (section == 0 || oldHead == chainIndexer.getSectionHead(section-1)) {
					chainIndexer.setSectionHead(section, newHead)
					chainIndexer.setValidStoredSections(section + 1)
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

		if chainIndexer.getSectionHead(chainIndexer.storedSections-1) == hash {
			return
		}
		chainIndexer.setValidStoredSections(chainIndexer.storedSections - 1)
	}
}

// processSection processes an entire section by calling backend functions while
// ensuring the continuity of the passed headers. Since the chain mutex is not
// held while processing, the continuity can be broken by a long reorg, in which
// case the function returns with an error.
func (chainIndexer *ChainIndexerService) processSection(section uint64, lastHead crypto.Hash) (crypto.Hash, error) {
	log.WithField("section", section).Trace("Processing new chain section")
	// Reset and partial processing

	if err := chainIndexer.reset(chainIndexer.ctx, section, lastHead); err != nil {
		chainIndexer.setValidStoredSections(0)
		return crypto.Hash{}, err
	}

	for number := section * chainIndexer.Config.SectionSize; number < (section+1)*chainIndexer.Config.SectionSize; number++ {
		hash := crypto.Hash{}
		blockHeader, err := chainIndexer.ChainService.GetBlockHeaderByHeight(number)
		if err == nil {
			hash = *blockHeader.Hash()
		}

		if hash == (crypto.Hash{}) {
			return crypto.Hash{}, fmt.Errorf("canonical block #%d unknown", number)
		}

		if blockHeader == nil {
			return crypto.Hash{}, fmt.Errorf("block #%d [%x…] not found", number, hash[:4])
		} else if blockHeader.PreviousHash != lastHead {
			return crypto.Hash{}, fmt.Errorf("chain reorged during section processing")
		}
		if err := chainIndexer.process(chainIndexer.ctx, blockHeader); err != nil {
			return crypto.Hash{}, err
		}
		lastHead = *blockHeader.Hash()
	}
	if err := chainIndexer.commit(); err != nil {
		return crypto.Hash{}, err
	}
	return lastHead, nil
}


// setValidSections writes the number of valid sections to the index database
func (chainIndexer *ChainIndexerService) setValidStoredSections(sections uint64) {
	// Set the current number of valid sections in the database
	chainIndexer.setStoredSections(sections)

	// Remove any reorged sections, caching the valids in the mean time
	for chainIndexer.storedSections > sections {
		chainIndexer.storedSections--
		chainIndexer.deleteSectionHead(chainIndexer.storedSections)
	}
	chainIndexer.storedSections = sections // needed if new > old
}

// 启动新的bloombits索引部分。
func (chainIndexer *ChainIndexerService) reset(ctx context.Context, section uint64, lastSectionHead crypto.Hash) error {
	gen, err := bloombits.NewGenerator(uint(chainIndexer.size))
	chainIndexer.gen, chainIndexer.section, chainIndexer.head = gen, section, crypto.Hash{}
	return err
}

// 将新区块头的bloom添加到索引。
func (chainIndexer *ChainIndexerService) process(ctx context.Context, header *types.BlockHeader) error {
	chainIndexer.gen.AddBloom(uint(header.Height - chainIndexer.section * chainIndexer.size), header.Bloom)
	chainIndexer.head = *header.Hash()
	return nil
}

// 完成bloom部分和把它写进数据库。
func (chainIndexer *ChainIndexerService) commit() error {
	batch := chainIndexer.DatabaseService.NewBatch()
	for i := 0; i < types.BloomBitLength; i++ {
		bits, err := chainIndexer.gen.Bitset(uint(i))
		if err != nil {
			return err
		}
		if err := batch.Put(bloomBitsKey(uint(i), chainIndexer.section, chainIndexer.head), bitutil.CompressBytes(bits)) ; err != nil {
			log.WithField("err", err).Fatal("Failed to store bloom bits")
		}
	}
	return batch.Write()
}