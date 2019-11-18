package blockmgr

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/common/event"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/network/p2p"
	"github.com/drep-project/DREP-Chain/types"
	"gopkg.in/urfave/cli.v1"
)

type p2pServiceMock struct {
	app.Service
}

func (ps *p2pServiceMock) SendAsync(w p2p.MsgWriter, msgType uint64, msg interface{}) chan error {
	return nil
}
func (ps *p2pServiceMock) Send(w p2p.MsgWriter, msgType uint64, msg interface{}) error {
	return nil
}
func (ps *p2pServiceMock) Peers() []*p2p.Peer {
	return nil
}
func (ps *p2pServiceMock) AddPeer(nodeUrl string) error {
	return nil
}
func (ps *p2pServiceMock) RemovePeer(url string) {
}
func (ps *p2pServiceMock) AddProtocols(protocols []p2p.Protocol) {
}
func (ps *p2pServiceMock) Name() string {
	return ""
} // service  name must be unique
func (ps *p2pServiceMock) Api() []app.API {
	return nil
} // Interfaces required for services
func (ps *p2pServiceMock) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, nil
}
func (ps *p2pServiceMock) Init(executeContext *app.ExecuteContext) error {
	return nil
}
func (ps *p2pServiceMock) Start(executeContext *app.ExecuteContext) error {
	return nil
}
func (ps *p2pServiceMock) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

type peerInfoMock struct {
	height uint64
}

func (p *peerInfoMock) GetMsgRW() p2p.MsgReadWriter {
	return nil
}

func (p *peerInfoMock) GetHeight() uint64 {
	return p.height
}

func (p *peerInfoMock) GetAddr() string {
	return "127.0.0.1"
}

func (p *peerInfoMock) SetHeight(height uint64) {
	p.height = height
}
func (p *peerInfoMock) KnownTx(tx *types.Transaction) bool {
	return true
}
func (p *peerInfoMock) MarkTx(tx *types.Transaction) {

}
func (p *peerInfoMock) KnownBlock(blk *types.Block) bool {
	return true
}
func (p *peerInfoMock) MarkBlock(blk *types.Block) {}

//var pi types.PeerInfoInterface = &peerInfoMock{}

type chainServiceMock struct {
}

func (ps *chainServiceMock) Name() string {
	return ""
} // service  name must be unique
func (ps *chainServiceMock) Api() []app.API {
	return nil
} // Interfaces required for services
func (ps *chainServiceMock) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, nil
}
func (ps *chainServiceMock) Init(executeContext *app.ExecuteContext) error {
	return nil
}
func (ps *chainServiceMock) Start(executeContext *app.ExecuteContext) error {
	return nil
}
func (ps *chainServiceMock) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (ps *chainServiceMock) ChainID() app.ChainIdType {
	return [64]byte{}
}

func (ps *chainServiceMock) DeriveMerkleRoot(txs []*types.Transaction) []byte {
	return nil
}
func (ps *chainServiceMock) GetBlockByHash(hash *crypto.Hash) (*types.Block, error) {
	return nil, nil
}
func (ps *chainServiceMock) GetBlockByHeight(number uint64) (*types.Block, error) {
	return nil, nil
}

//DefaultChainConfig
func (ps *chainServiceMock) GetBlockHeaderByHash(hash *crypto.Hash) (*types.BlockHeader, error) {
	return nil, nil
}
func (ps *chainServiceMock) GetBlockHeaderByHeight(number uint64) (*types.BlockHeader, error) {
	return nil, nil
}
func (ps *chainServiceMock) GetBlocksFrom(start, size uint64) ([]*types.Block, error) {
	return nil, nil
}

func (ps *chainServiceMock) GetCurrentState() *database.Database {
	return nil
}
func (ps *chainServiceMock) GetHeader(hash crypto.Hash, number uint64) *types.BlockHeader {
	return nil
}
func (ps *chainServiceMock) GetHighestBlock() (*types.Block, error) {
	return nil, nil
}
func (ps *chainServiceMock) RootChain() app.ChainIdType {
	return [64]byte{}
}

func (ps *chainServiceMock) BestChain() *chain.ChainView {
	cv := &chain.ChainView{}
	return cv
}
func (ps *chainServiceMock) CalcGasLimit(parent *types.BlockHeader, gasFloor, gasCeil uint64) *big.Int {
	return nil
}

func (ps *chainServiceMock) ProcessBlock(block *types.Block) (bool, bool, error) {
	return true, true, nil
}

func (ps *chainServiceMock) NewBlockFeed() *event.Feed {
	return nil
}
func (ps *chainServiceMock) BlockExists(blockHash *crypto.Hash) bool {
	return true
}
func (ps *chainServiceMock) TransactionValidator() chain.ITransactionValidator {
	return nil
}
func (ps *chainServiceMock) GetDatabaseService() *database.DatabaseService {
	return nil
}
func (ps *chainServiceMock) Index() *chain.BlockIndex {
	return nil
}
func (ps *chainServiceMock) BlockValidator() chain.IBlockValidator {
	return nil
}
func (ps *chainServiceMock) Config() *types.ChainConfig {
	return nil
}
func (ps *chainServiceMock) AccumulateRewards(db *database.Database, b *types.Block, totalGasBalance *big.Int) error {
	return nil
}
func (ps *chainServiceMock) DetachBlockFeed() *event.Feed {
	return nil
}

//var bm *BlockMgr

func prepareBase(t *testing.T) (*BlockMgr, []*types.Block) {
	db, err := database.NewDatabase("./test/")
	if err != nil {
		t.Fatal(err)
	}
	blks := generatorChain(t, db)

	ds := database.NewDatabaseService(db)
	cs := chain.NewChainService(chain.DefaultChainConfig, ds)

	bm := NewBlockMgr(DefaultChainConfig, "./", cs, &p2pServiceMock{})

	return bm, blks
}

func TestFindAncestor(t *testing.T) {
	bm, blks := prepareBase(t)

	peerInfo := &peerInfoMock{}
	headerHashs := []*syncHeaderHash{}

	blks = blks[9:]
	for _, b := range blks {
		headerHashs = append(headerHashs, &syncHeaderHash{headerHash: b.Header.Hash(), height: b.Header.Height})
	}

	//超时错误
	ancestor, err := bm.findAncestor(peerInfo)
	if err == nil {
		t.Fatal(err)
	}

	go func() {
		bm.headerHashCh <- headerHashs
	}()
	ancestor, err = bm.findAncestor(peerInfo)
	if err != nil {
		t.Fatal(err)
	}

	if ancestor != 10 {
		t.Fatal("get ancestor err", "need:", 0, "ancestor：", ancestor)
	}

	os.RemoveAll("./test")
}

func TestFetchBlocks(t *testing.T) {

	defer os.RemoveAll("./test")

	peer := &peerInfoMock{}
	headerHashs1 := []*syncHeaderHash{}
	headerHashs2 := []*syncHeaderHash{}

	bm, blks := prepareBase(t)

	blks1 := blks[1:2]
	for _, b := range blks1 {
		headerHashs1 = append(headerHashs1, &syncHeaderHash{headerHash: b.Header.Hash(), height: b.Header.Height})
	}

	go func() {
		time.Sleep(time.Second * 4)
		bm.headerHashCh <- headerHashs1
	}()

	//fake header hash
	blks2 := blks[2:4]
	for _, b := range blks2 {
		headerHashs2 = append(headerHashs2, &syncHeaderHash{headerHash: b.Header.Hash(), height: b.Header.Height})
	}

	go func() {
		time.Sleep(time.Second * 8)
		bm.headerHashCh <- headerHashs2
	}()

	//fake block body
	go func() {
		time.Sleep(time.Second * 12)
		bm.blocksCh <- blks[2:4]
	}()

	peer.height = 4
	bm.peersInfo["127.0.0.1"] = peer
	err := bm.fetchBlocks(peer)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClearSyncCh(t *testing.T) {
	//clearSyncCh()
	//select {
	//case <-blockMgr.headerHashCh:
	//default:
	//}
	//
	//select {
	//case <-blockMgr.blocksCh:
	//default:
	//}
	//
	//select {
	//case <-blockMgr.syncTimerCh:
	//default:
	//}
	//
	//blockMgr.allTasks = newHeightSortedMap()
	//
	//blockMgr.pendingSyncTasks.Range(func(key, value interface{}) bool {
	//	blockMgr.pendingSyncTasks.Delete(key)
	//	return true
	//})
}
