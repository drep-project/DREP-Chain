package blockmgr

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/service/chainservice"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/network/p2p"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"os"
	"testing"
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
}

func (p *peerInfoMock) GetMsgRW() p2p.MsgReadWriter {
	return nil
}

func (p *peerInfoMock) GetHeight() uint64 {
	return 0
}

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

func (ps *chainServiceMock) DeriveMerkleRoot(txs []*chainTypes.Transaction) []byte {
	return nil
}
func (ps *chainServiceMock) GetBlockByHash(hash *crypto.Hash) (*chainTypes.Block, error) {
	return nil, nil
}
func (ps *chainServiceMock) GetBlockByHeight(number uint64) (*chainTypes.Block, error) {
	return nil, nil
}

//DefaultChainConfig
func (ps *chainServiceMock) GetBlockHeaderByHash(hash *crypto.Hash) (*chainTypes.BlockHeader, error) {
	return nil, nil
}
func (ps *chainServiceMock) GetBlockHeaderByHeight(number uint64) (*chainTypes.BlockHeader, error) {
	return nil, nil
}
func (ps *chainServiceMock) GetBlocksFrom(start, size uint64) ([]*chainTypes.Block, error) {
	return nil, nil
}

func (ps *chainServiceMock) GetCurrentState() *database.Database {
	return nil
}
func (ps *chainServiceMock) GetHeader(hash crypto.Hash, number uint64) *chainTypes.BlockHeader {
	return nil
}
func (ps *chainServiceMock) GetHighestBlock() (*chainTypes.Block, error) {
	return nil, nil
}
func (ps *chainServiceMock) RootChain() app.ChainIdType {
	return [64]byte{}
}

func (ps *chainServiceMock) BestChain() *chainservice.ChainView {
	cv := &chainservice.ChainView{}
	return cv
}
func (ps *chainServiceMock) CalcGasLimit(parent *chainTypes.BlockHeader, gasFloor, gasCeil uint64) *big.Int {
	return nil
}

func (ps *chainServiceMock) ProcessBlock(block *chainTypes.Block) (bool, bool, error) {
	return true, true, nil
}

func (ps *chainServiceMock) NewBlockFeed() *event.Feed {
	return nil
}
func (ps *chainServiceMock) BlockExists(blockHash *crypto.Hash) bool {
	return true
}
func (ps *chainServiceMock) TransactionValidator() chainservice.ITransactionValidator {
	return nil
}
func (ps *chainServiceMock) GetDatabaseService() *database.DatabaseService {
	return nil
}
func (ps *chainServiceMock) Index() *chainservice.BlockIndex {
	return nil
}
func (ps *chainServiceMock) BlockValidator() chainservice.IBlockValidator {
	return nil
}
func (ps *chainServiceMock) Config() *chainTypes.ChainConfig {
	return nil
}
func (ps *chainServiceMock) AccumulateRewards(db *database.Database, b *chainTypes.Block, totalGasBalance *big.Int) error {
	return nil
}
func (ps *chainServiceMock) DetachBlockFeed() *event.Feed {
	return nil
}

var bm *BlockMgr

func TestFindAncestor(t *testing.T) {

	db, err := database.NewDatabase("./test/")
	if err != nil {
		t.Fatal(err)
	}
	blks := generatorChain(t, db)

	ds := database.NewDatabaseService(db)
	cs := chainservice.NewChainService(chainservice.DefaultChainConfig, ds)

	bm = NewBlockMgr(DefaultChainConfig, "./", cs, &p2pServiceMock{})

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

func testFetchBlocks(t *testing.T) {
	peer := &peerInfoMock{}

	err := bm.fetchBlocks(peer)
	if err != nil {
		t.Fatal(err)
	}
}

func testClearSyncCh() {
	//clearSyncCh()
}
