package chain_indexer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/common/bitutil"
	"github.com/drep-project/drep-chain/common/bloombits"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/types"
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
	BloomStatus() (uint64, uint64)
	GetConfig() *ChainIndexerConfig
	GetIndexerStore() *ChainIndexerStore
}

var _ ChainIndexerServiceInterface = &ChainIndexerService{}

type ChainIndexerService struct {
	DatabaseService *database.DatabaseService   `service:"database"`
	ChainService    chain.ChainServiceInterface `service:"chain"`
	Config          *ChainIndexerConfig

	db             *ChainIndexerStore
	storedSections uint64 // Number of sections successfully indexed into the database
	knownSections  uint64 // Number of sections known to be complete (block wise)

	active    uint32          // Flag whether the event loop was started
	update    chan struct{}   // Notification channel that headers should be processed
	quit      chan chan error // Quit channel to tear down running goroutines
	ctx       context.Context
	ctxCancel func()

	lock sync.RWMutex

	// bloomIndexer
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
	// initialize module config
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
	chainIndexer.db = NewChainIndexerStore(chainIndexer.DatabaseService.LevelDb())
	chainIndexer.storedSections = chainIndexer.db.getStoredSections()

	go chainIndexer.updateLoop()

	return nil
}

func (chainIndexer *ChainIndexerService) Start(executeContext *app.ExecuteContext) error {
	events := make(chan *types.ChainEvent, 10)
	sub := chainIndexer.ChainService.NewBlockFeed().Subscribe(events)

	go chainIndexer.eventLoop(chainIndexer.ChainService.GetCurrentHeader(), events, sub)

	return nil
}

func (chainIndexer *ChainIndexerService) Stop(executeContext *app.ExecuteContext) error {
	var errs []error
	chainIndexer.ctxCancel()

	// Tear down the primary update loop
	errc := make(chan error)
	chainIndexer.quit <- errc
	if err := <-errc; err != nil {
		errs = append(errs, err)
	}
	// If needed, tear down the secondary event loop
	if atomic.LoadUint32(&chainIndexer.active) != 0 {
		chainIndexer.quit <- errc
		if err := <-errc; err != nil {
			errs = append(errs, err)
		}
	}
	// Return any failures
	switch {
	case len(errs) == 0:
		return nil

	case len(errs) == 1:
		return errs[0]

	default:
		return fmt.Errorf("%v", errs)
	}

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
						log.WithField("percentage", chainIndexer.storedSections*100/chainIndexer.knownSections).Info("Upgrading chain index")
					}
					updated = time.Now()
				}
				// Cache the current section count and head to allow unlocking the mutex
				chainIndexer.verifyLastHead()
				section := chainIndexer.storedSections
				var oldHead crypto.Hash
				if section > 0 {
					oldHead = chainIndexer.db.getSectionHead(section - 1)
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
				if err == nil && (section == 0 || oldHead == chainIndexer.db.getSectionHead(section-1)) {
					chainIndexer.db.setSectionHead(section, newHead)
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

// eventLoop is a secondary - optional - event loop of the indexer which is only
// started for the outermost indexer to push chain head events into a processing
// queue.
func (chainIndexer *ChainIndexerService) eventLoop(currentHeader *types.BlockHeader, events chan *types.ChainEvent, sub event.Subscription) {
	// Mark the chain indexer as active, requiring an additional teardown
	atomic.StoreUint32(&chainIndexer.active, 1)

	defer sub.Unsubscribe()

	// Fire the initial new head event to start any outstanding processing
	chainIndexer.newHead(currentHeader.Height, false)

	var (
		prevHeader = currentHeader
		prevHash   = *currentHeader.Hash()
	)
	for {
		select {
		case errc := <-chainIndexer.quit:
			// Chain indexer terminating, report no failure and abort
			errc <- nil
			return

		case block, ok := <-events:
			// Received a new event, ensure it's not nil (closing) and update
			if !ok {
				errc := <-chainIndexer.quit
				errc <- nil
				return
			}
			header := block.Block.Header
			if header.PreviousHash != prevHash {
				// Reorg to the common ancestor if needed (might not exist in light sync mode, skip reorg then)
				// TODO(karalabe, zsfelfoldi): This seems a bit brittle, can we detect this case explicitly?

				hash := crypto.Hash{}
				blockHeader, err := chainIndexer.ChainService.GetBlockHeaderByHeight(prevHeader.Height)
				if err == nil {
					hash = *blockHeader.Hash()
				}

				if hash != prevHash {
					if h := chainIndexer.db.FindCommonAncestor(prevHeader, header); h != nil {
						chainIndexer.newHead(h.Height, true)
					}
				}
			}
			chainIndexer.newHead(header.Height, false)

			prevHeader, prevHash = header, *header.Hash()
		}
	}
}

// newHead notifies the indexer about new chain heads and/or reorgs.
func (chainIndexer *ChainIndexerService) newHead(head uint64, reorg bool) {
	chainIndexer.lock.Lock()
	defer chainIndexer.lock.Unlock()

	// If a reorg happened, invalidate all sections until that point
	if reorg {
		// Revert the known section number to the reorg point
		known := (head + 1) / chainIndexer.Config.SectionSize
		stored := known

		if known < chainIndexer.knownSections {
			chainIndexer.knownSections = known
		}
		// Revert the stored sections from the database to the reorg point
		if stored < chainIndexer.storedSections {
			chainIndexer.setValidStoredSections(stored)
		}
		// Update the new head number to the finalized section end and notify children
		head = known * chainIndexer.Config.SectionSize

		return
	}
	// No reorg, calculate the number of newly known sections and update if high enough
	var sections uint64
	if head >= chainIndexer.Config.ConfirmsReq {
		sections = (head + 1 - chainIndexer.Config.ConfirmsReq) / chainIndexer.Config.SectionSize

		if sections > chainIndexer.knownSections {

			chainIndexer.knownSections = sections

			select {
			case chainIndexer.update <- struct{}{}:
			default:
			}
		}
	}
}

// verifyLastHead compares last stored section head with the corresponding block hash in the
// actual canonical chain and rolls back reorged sections if necessary to ensure that stored
// sections are all valid
func (chainIndexer *ChainIndexerService) verifyLastHead() {
	for chainIndexer.storedSections > 0 {

		hash := crypto.Hash{}
		blockHeader, err := chainIndexer.ChainService.GetBlockHeaderByHeight(chainIndexer.storedSections*chainIndexer.Config.SectionSize - 1)
		if err == nil {
			hash = *blockHeader.Hash()
		}

		if chainIndexer.db.getSectionHead(chainIndexer.storedSections-1) == hash {
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

// 启动新的bloombits索引部分。
func (chainIndexer *ChainIndexerService) reset(ctx context.Context, section uint64, lastSectionHead crypto.Hash) error {
	gen, err := bloombits.NewGenerator(uint(chainIndexer.Config.SectionSize))
	chainIndexer.gen, chainIndexer.section, chainIndexer.head = gen, section, crypto.Hash{}
	return err
}

// 将新区块头的bloom添加到索引。
func (chainIndexer *ChainIndexerService) process(ctx context.Context, header *types.BlockHeader) error {
	chainIndexer.gen.AddBloom(uint(header.Height-chainIndexer.section*chainIndexer.Config.SectionSize), header.Bloom)
	chainIndexer.head = *header.Hash()
	return nil
}

// 完成bloom部分和把它写进数据库。
func (chainIndexer *ChainIndexerService) commit() error {
	batch := chainIndexer.db.NewBatch()
	for i := 0; i < types.BloomBitLength; i++ {
		bits, err := chainIndexer.gen.Bitset(uint(i))
		if err != nil {
			return err
		}
		if err := batch.Put(bloomBitsKey(uint(i), chainIndexer.section, chainIndexer.head), bitutil.CompressBytes(bits)); err != nil {
			log.WithField("err", err).Fatal("Failed to store bloom bits")
		}
	}
	return batch.Write()
}

// setValidStoredSections writes the number of valid sections to the index database
func (chainIndexer *ChainIndexerService) setValidStoredSections(sections uint64) {
	// Set the current number of valid sections in the database
	chainIndexer.db.setStoredSections(sections)

	// Remove any reorged sections, caching the valids in the mean time
	for chainIndexer.storedSections > sections {
		chainIndexer.storedSections--
		chainIndexer.db.deleteSectionHead(chainIndexer.storedSections)
	}
	chainIndexer.storedSections = sections // needed if new > old
}

// Sections returns the number of processed sections maintained by the indexer
// and also the information about the last header indexed for potential canonical
// verifications.
func (chainIndexer *ChainIndexerService) Sections() (uint64, uint64, crypto.Hash) {
	chainIndexer.lock.Lock()
	defer chainIndexer.lock.Unlock()

	chainIndexer.verifyLastHead()
	return chainIndexer.storedSections, chainIndexer.storedSections*chainIndexer.Config.SectionSize - 1, chainIndexer.db.getSectionHead(chainIndexer.storedSections - 1)
}

func (chainIndexer *ChainIndexerService) BloomStatus() (uint64, uint64) {
	sections, _, _ := chainIndexer.Sections()
	return chainIndexer.Config.SectionSize, sections
}

func (chainIndexer *ChainIndexerService) GetConfig() *ChainIndexerConfig {
	return chainIndexer.Config
}

func (chainIndexer *ChainIndexerService) GetIndexerStore() *ChainIndexerStore {
	return chainIndexer.db
}
func (chainIndexer *ChainIndexerService) DefaultConfig() *ChainIndexerConfig {
	return DefaultConfig
}
