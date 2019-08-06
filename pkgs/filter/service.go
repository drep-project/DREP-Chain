package filter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/blockmgr"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/bitutil"
	"github.com/drep-project/drep-chain/common/bloombits"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/pkgs/chain_indexer"
	"github.com/drep-project/drep-chain/types"
)

var (
	deadline = 5 * time.Minute // consider a filter inactive if it has not been polled for within deadline

	// bloomServiceThreads is the number of goroutines used globally by an Ethereum
	// instance to service bloombits lookups for all running filters.
	bloomServiceThreads = 16

	// bloomFilterThreads is the number of goroutines used locally per filter to
	// multiplex requests onto the global servicing goroutines.
	bloomFilterThreads = 3

	// bloomRetrievalBatch is the maximum number of bloom bit retrievals to service
	// in a single batch.
	bloomRetrievalBatch = 16

	// bloomRetrievalWait is the maximum time to wait for enough bloom bit requests
	// to accumulate request an entire batch (avoiding hysteresis).
	bloomRetrievalWait = time.Duration(0)
)

type ServiceDatabase interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

type Backend interface {
	HeaderByNumber(ctx context.Context, blockNr common.BlockNumber) (*types.BlockHeader, error)
	HeaderByHash(ctx context.Context, blockHash crypto.Hash) (*types.BlockHeader, error)
	GetReceipts(ctx context.Context, blockHash crypto.Hash) (types.Receipts, error)
	GetLogsByHash(ctx context.Context, blockHash crypto.Hash) ([][]*types.Log, error)

	SubscribeNewTxsEvent(chan<- types.NewTxsEvent) event.Subscription
	SubscribeChainEvent(ch chan<- types.ChainEvent) event.Subscription
	SubscribeRemovedLogsEvent(ch chan<- types.RemovedLogsEvent) event.Subscription
	SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription

	BloomStatus() (uint64, uint64)
	ServiceFilter(ctx context.Context, session *bloombits.MatcherSession)
}

type FilterServiceInterface interface {
	app.Service
	Backend
}

var _ FilterServiceInterface = &FilterService{}

type FilterService struct {
	DatabaseService 		*database.DatabaseService		`service:"database"`
	Blockmgr        		*blockmgr.BlockMgr        		`service:"blockmgr"`
	ChainService    		chain.ChainServiceInterface		`service:"chain"`
	ChainIndexerService		chain_indexer.ChainIndexerServiceInterface `service:"chain_indexer"`
	Config					*FilterConfig

	apis            		[]app.API

	bloomRequests			chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	mux       				*event.TypeMux
	events    				*EventSystem
	filtersMu 				sync.Mutex
	filters   				map[ID]*filter
}

// filter is a helper struct that holds meta information over the filter type
// and associated subscription in the event system.
type filter struct {
	typ      Type
	deadline *time.Timer // filter is inactiv when deadline triggers
	hashes   []crypto.Hash
	crit     FilterQuery
	logs     []*types.Log
	s        *Subscription // associated subscription in event system
}

// implement Service interface
func (service *FilterService) Name() string {
	return MODULENAME
}

func (service *FilterService) Api() []app.API {
	return service.apis
}

func (service *FilterService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{EnableFilterFlag}
}

func (service *FilterService) Init(executeContext *app.ExecuteContext) error {
	// check service dependencies
	if service.DatabaseService == nil {
		return fmt.Errorf("batabaseService not init")
	}
	if service.Blockmgr == nil {
		return fmt.Errorf("batabaseService not init")
	}
	if service.ChainService == nil {
		return fmt.Errorf("chainService not init")
	}
	if service.ChainIndexerService == nil {
		return fmt.Errorf("chainIndexerService not init")
	}

	// initialize module config
	service.Config = DefaultConfig
	err := executeContext.UnmashalConfig(service.Name(), service.Config)
	if err != nil {
		return err
	}
	if executeContext.Cli.GlobalIsSet(EnableFilterFlag.Name) {
		service.Config.Enable = executeContext.Cli.GlobalBool(EnableFilterFlag.Name)
	}
	if !service.Config.Enable {
		return nil
	}

	// initialize other fields in service
	service.mux = new(event.TypeMux)
	service.events = NewEventSystem(service.mux, service, false)
	service.filters = make(map[ID]*filter)
	service.bloomRequests = make(chan chan *bloombits.Retrieval)

	service.apis = []app.API{
		app.API{
			Namespace: MODULENAME,
			Version:   "1.0",
			Service: &FilterApi{
				filterService: service,
			},
			Public: true,
		},
	}

	go service.timeoutLoop()

	return nil
}

func (service *FilterService) Start(executeContext *app.ExecuteContext) error {
	service.startBloomHandlers(service.ChainIndexerService.GetConfig().SectionSize)

	return nil
}

func (service *FilterService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

// ------------------------------------
// implement service logic

// startBloomHandlers starts a batch of goroutines to accept bloom bit database
// retrievals from possibly a range of filters and serving the data to satisfy.
func (service *FilterService) startBloomHandlers(sectionSize uint64) {
	for i := 0; i < bloomServiceThreads; i++ {
		go func() {
			for {
				request := <- service.bloomRequests
				task := <-request
				task.Bitsets = make([][]byte, len(task.Sections))
				for i, section := range task.Sections {

					head := crypto.Hash{}
					blockHeader, err := service.ChainService.GetBlockHeaderByHeight((section+1)*sectionSize-1)
					if err == nil {
						head = *blockHeader.Hash()
					}

					if compVector, err := service.ChainIndexerService.ReadBloomBits(task.Bit, section, head); err == nil {
						if blob, err := bitutil.DecompressBytes(compVector, int(sectionSize/8)); err == nil {
							task.Bitsets[i] = blob
						} else {
							task.Error = err
						}
					} else {
						task.Error = err
					}
				}
				request <- task
			}
		}()
	}
}

// timeoutLoop runs every 5 minutes and deletes filters that have not been recently used.
// Tt is started when the api is created.
func (service *FilterService) timeoutLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		<-ticker.C
		service.filtersMu.Lock()
		for id, f := range service.filters {
			select {
			case <-f.deadline.C:
				f.s.Unsubscribe()
				delete(service.filters, id)
			default:
				continue
			}
		}
		service.filtersMu.Unlock()
	}
}

// NewPendingTransactionFilter creates a filter that fetches pending transaction hashes
// as transactions enter the pending state.
//
// It is part of the filter package because this filter can be used through the
// `eth_getFilterChanges` polling method that is also used for log filters.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_newpendingtransactionfilter
func (service *FilterService) NewPendingTransactionFilter() ID {
	var (
		pendingTxs   = make(chan []crypto.Hash)
		pendingTxSub = service.events.SubscribePendingTxs(pendingTxs)
	)

	service.filtersMu.Lock()
	service.filters[pendingTxSub.ID] = &filter{typ: PendingTransactionsSubscription, deadline: time.NewTimer(deadline), hashes: make([]crypto.Hash, 0), s: pendingTxSub}
	service.filtersMu.Unlock()

	go func() {
		for {
			select {
			case ph := <-pendingTxs:
				service.filtersMu.Lock()
				if f, found := service.filters[pendingTxSub.ID]; found {
					f.hashes = append(f.hashes, ph...)
				}
				service.filtersMu.Unlock()
			case <-pendingTxSub.Err():
				service.filtersMu.Lock()
				delete(service.filters, pendingTxSub.ID)
				service.filtersMu.Unlock()
				return
			}
		}
	}()

	return pendingTxSub.ID
}

// NewBlockFilter creates a filter that fetches blocks that are imported into the chain.
// It is part of the filter package since polling goes with eth_getFilterChanges.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_newblockfilter
func (service *FilterService) NewBlockFilter() ID {
	var (
		headers   = make(chan *types.BlockHeader)
		headerSub = service.events.SubscribeNewHeads(headers)
	)

	service.filtersMu.Lock()
	service.filters[headerSub.ID] = &filter{typ: BlocksSubscription, deadline: time.NewTimer(deadline), hashes: make([]crypto.Hash, 0), s: headerSub}
	service.filtersMu.Unlock()

	go func() {
		for {
			select {
			case h := <-headers:
				service.filtersMu.Lock()
				if f, found := service.filters[headerSub.ID]; found {
					f.hashes = append(f.hashes, *h.Hash())
				}
				service.filtersMu.Unlock()
			case <-headerSub.Err():
				service.filtersMu.Lock()
				delete(service.filters, headerSub.ID)
				service.filtersMu.Unlock()
				return
			}
		}
	}()

	return headerSub.ID
}

// NewFilter creates a new filter and returns the filter id. It can be
// used to retrieve logs when the state changes. This method cannot be
// used to fetch logs that are already stored in the state.
//
// Default criteria for the from and to block are "latest".
// Using "latest" as block number will return logs for mined blocks.
// Using "pending" as block number returns logs for not yet mined (pending) blocks.
// In case logs are removed (chain reorg) previously returned logs are returned
// again but with the removed property set to true.
//
// In case "fromBlock" > "toBlock" an error is returned.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_newfilter
func (service *FilterService) NewFilter(crit FilterQuery) (ID, error) {
	logs := make(chan []*types.Log)
	logsSub, err := service.events.SubscribeLogs(crit, logs)
	if err != nil {
		return ID(""), err
	}

	service.filtersMu.Lock()
	service.filters[logsSub.ID] = &filter{typ: LogsSubscription, crit: crit, deadline: time.NewTimer(deadline), logs: make([]*types.Log, 0), s: logsSub}
	service.filtersMu.Unlock()

	go func() {
		for {
			select {
			case l := <-logs:
				service.filtersMu.Lock()
				if f, found := service.filters[logsSub.ID]; found {
					f.logs = append(f.logs, l...)
				}
				service.filtersMu.Unlock()
			case <-logsSub.Err():
				service.filtersMu.Lock()
				delete(service.filters, logsSub.ID)
				service.filtersMu.Unlock()
				return
			}
		}
	}()

	return logsSub.ID, nil
}

// GetLogs returns logs matching the given argument that are stored within the state.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getlogs
func (service *FilterService) GetLogs(ctx context.Context, crit FilterQuery) ([]*types.Log, error) {
	var filter *Filter
	if crit.BlockHash != nil {
		// Block filter requested, construct a single-shot filter
		filter = NewBlockFilter(service, *crit.BlockHash, crit.Addresses, crit.Topics)
	} else {
		// Convert the RPC block numbers into internal representations
		begin := common.LatestBlockNumber.Int64()
		if crit.FromBlock != nil {
			begin = crit.FromBlock.Int64()
		}
		end := common.LatestBlockNumber.Int64()
		if crit.ToBlock != nil {
			end = crit.ToBlock.Int64()
		}
		// Construct the range filter
		filter = NewRangeFilter(service, begin, end, crit.Addresses, crit.Topics)
	}
	// Run the filter and return all the logs
	logs, err := filter.Logs(ctx)
	if err != nil {
		return nil, err
	}
	return returnLogs(logs), err
}

// UninstallFilter removes the filter with the given filter id.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_uninstallfilter
func (service *FilterService) UninstallFilter(id ID) bool {
	service.filtersMu.Lock()
	f, found := service.filters[id]
	if found {
		delete(service.filters, id)
	}
	service.filtersMu.Unlock()
	if found {
		f.s.Unsubscribe()
	}

	return found
}

// GetFilterLogs returns the logs for the filter with the given id.
// If the filter could not be found an empty array of logs is returned.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getfilterlogs
func (service *FilterService) GetFilterLogs(ctx context.Context, id ID) ([]*types.Log, error) {
	service.filtersMu.Lock()
	f, found := service.filters[id]
	service.filtersMu.Unlock()

	if !found || f.typ != LogsSubscription {
		return nil, fmt.Errorf("filter not found")
	}

	var filter *Filter
	if f.crit.BlockHash != nil {
		// Block filter requested, construct a single-shot filter
		filter = NewBlockFilter(service, *f.crit.BlockHash, f.crit.Addresses, f.crit.Topics)
	} else {
		// Convert the RPC block numbers into internal representations
		begin := common.LatestBlockNumber.Int64()
		if f.crit.FromBlock != nil {
			begin = f.crit.FromBlock.Int64()
		}
		end := common.LatestBlockNumber.Int64()
		if f.crit.ToBlock != nil {
			end = f.crit.ToBlock.Int64()
		}
		// Construct the range filter
		filter = NewRangeFilter(service, begin, end, f.crit.Addresses, f.crit.Topics)
	}
	// Run the filter and return all the logs
	logs, err := filter.Logs(ctx)
	if err != nil {
		return nil, err
	}
	return returnLogs(logs), nil
}

// GetFilterChanges returns the logs for the filter with the given id since
// last time it was called. This can be used for polling.
//
// For pending transaction and block filters the result is []common.Hash.
// (pending)Log filters return []Log.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getfilterchanges
func (service *FilterService) GetFilterChanges(id ID) (interface{}, error) {
	service.filtersMu.Lock()
	defer service.filtersMu.Unlock()

	if f, found := service.filters[id]; found {
		if !f.deadline.Stop() {
			// timer expired but filter is not yet removed in timeout loop
			// receive timer value and reset timer
			<-f.deadline.C
		}
		f.deadline.Reset(deadline)

		switch f.typ {
		case PendingTransactionsSubscription, BlocksSubscription:
			hashes := f.hashes
			f.hashes = nil
			return returnHashes(hashes), nil
		case LogsSubscription:
			logs := f.logs
			f.logs = nil
			return returnLogs(logs), nil
		}
	}

	return []interface{}{}, fmt.Errorf("filter not found")
}

// returnHashes is a helper that will return an empty hash array case the given hash array is nil,
// otherwise the given hashes array is returned.
func returnHashes(hashes []crypto.Hash) []crypto.Hash {
	if hashes == nil {
		return []crypto.Hash{}
	}
	return hashes
}

// returnLogs is a helper that will return an empty log array in case the given logs array is nil,
// otherwise the given logs array is returned.
func returnLogs(logs []*types.Log) []*types.Log {
	if logs == nil {
		return []*types.Log{}
	}
	return logs
}

// ------------------------------------
// implement Backend interface

func (service *FilterService) HeaderByNumber(ctx context.Context, blockNr common.BlockNumber) (*types.BlockHeader, error) {
	if blockNr == common.LatestBlockNumber {
		return service.ChainService.GetCurrentHeader(), nil
	}
	return service.ChainService.GetBlockHeaderByHeight(uint64(blockNr.Int64()))
}

func (service *FilterService) HeaderByHash(ctx context.Context, blockHash crypto.Hash) (*types.BlockHeader, error) {
	return service.ChainService.GetBlockHeaderByHash(&blockHash)
}

func (service *FilterService) GetReceipts(ctx context.Context, blockHash crypto.Hash) (types.Receipts, error) {
	return service.DatabaseService.GetReceipts(blockHash), nil
}

func (service *FilterService) GetLogsByHash(ctx context.Context, blockHash crypto.Hash) ([][]*types.Log, error) {
	receipts := service.DatabaseService.GetReceipts(blockHash)
	if receipts == nil {
		return nil, nil
	}

	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (service *FilterService) SubscribeNewTxsEvent(ch chan<- types.NewTxsEvent) event.Subscription {
	return service.Blockmgr.NewTxFeed().Subscribe(ch)
}

func (service *FilterService) SubscribeChainEvent(ch chan<- types.ChainEvent) event.Subscription {
	return service.ChainService.NewBlockFeed().Subscribe(ch)
}

func (service *FilterService) SubscribeRemovedLogsEvent(ch chan<- types.RemovedLogsEvent) event.Subscription {
	return service.ChainService.GetRMLogsFeed().Subscribe(ch)
}

func (service *FilterService) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return service.ChainService.GetLogsFeed().Subscribe(ch)
}

func (service *FilterService) BloomStatus() (uint64, uint64) {
	return service.ChainIndexerService.BloomStatus()
}

func (service *FilterService) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, service.bloomRequests)
	}
}