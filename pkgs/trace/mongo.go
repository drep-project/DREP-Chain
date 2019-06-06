package trace

import (
	"context"
	"time"

	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongogDbStore used to save tx in mongo db, db name is "drep", col name is "tx"
type MongogDbStore struct {
	url       string
	client    *mongo.Client
	db        *mongo.Database
	txCol     *mongo.Collection
	blockCol *mongo.Collection
	headerCol *mongo.Collection

	viewTxCol     *mongo.Collection
	viewBlockCol *mongo.Collection
	viewHeaderCol *mongo.Collection
}

// NewMongogDbStore open a new db from url, if db not exist, auto create
func NewMongogDbStore(url string) (*MongogDbStore, error) {
	store := &MongogDbStore{
		url: url,
	}
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	var err error
	store.client, err = mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
	 	return  nil, err
	}
	err = store.client.Ping(ctx, nil)
	if err != nil {
		return  nil, err
	}
	store.db = store.client.Database("drep")
	store.txCol = store.db.Collection("tx")
	store.blockCol = store.db.Collection("block")
	store.headerCol = store.db.Collection("header")

	store.viewTxCol = store.db.Collection("view_tx")
	store.viewBlockCol = store.db.Collection("view_block")
	store.viewHeaderCol = store.db.Collection("view_header")
	return store, nil
}

func (store *MongogDbStore) InsertRecord(block *chainTypes.Block) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	rpcTxs := make([]interface{}, block.Data.TxCount)
	rpcHeader := chainTypes.RpcBlockHeader{}
	rpcHeader.FromBlockHeader(block.Header)
	rpcBlock := &chainTypes.RpcBlock{}
	rpcBlock.From(block)
	store.blockCol.InsertOne(ctx, rpcBlock)
	store.headerCol.InsertOne(ctx, rpcHeader)

	viewBlock := ViewBlock{}
	viewBlock.From(block)
	store.viewBlockCol.InsertOne(ctx, viewBlock)

	viewHeader := ViewBlockHeader{}
	viewHeader.From(block)
	store.viewHeaderCol.InsertOne(ctx, viewHeader)

	viewTxs := make([]interface{}, block.Data.TxCount)
	for index, tx := range block.Data.TxList {
		rpcTx := &chainTypes.RpcTransaction{}
		rpcTx.FromTx(tx)
		rpcTxs[index] = rpcTx

		viewTx := &ViewTransaction{}
		viewTx.FromTx(tx)
		viewTx.Height = block.Header.Height
		viewTxs[index] = viewTx
	}
	store.txCol.InsertMany(ctx, rpcTxs, nil)
	store.viewTxCol.InsertMany(ctx, viewTxs, nil)
}

func (store *MongogDbStore) DelRecord(block *chainTypes.Block) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	store.headerCol.DeleteOne(ctx, bson.M{"hash": block.Header.Hash()})
	store.blockCol.DeleteOne(ctx, bson.M{"hash": block.Header.Hash()})
	store.viewBlockCol.DeleteOne(ctx, bson.M{"hash": block.Header.Hash().String()})
	store.viewHeaderCol.DeleteOne(ctx, bson.M{"hash": block.Header.Hash().String()})
	for _, tx := range block.Data.TxList {
		store.txCol.DeleteOne(ctx, bson.M{"hash": tx.TxHash()})
		store.viewTxCol.DeleteOne(ctx, bson.M{"hash": tx.TxHash().String()})
	}
}

func (store *MongogDbStore) GetRawTransaction(txHash *crypto.Hash) ([]byte, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	curser, err := store.txCol.Find(ctx, bson.M{"hash": txHash})
	if err != nil {
		return nil, err
	}
	curser.Next(ctx)
	if curser.Current == nil {
		return nil, ErrTxNotFound
	}
	rpcTx := &chainTypes.RpcTransaction{}
	err = curser.Decode(rpcTx)
	if err != nil {
		return nil, err
	}
	tx := rpcTx.ToTx()
	return tx.AsPersistentMessage(), nil
}

func (store *MongogDbStore) GetTransaction(txHash *crypto.Hash) (*chainTypes.RpcTransaction, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	curser, err := store.txCol.Find(ctx, bson.M{"hash": txHash})
	if err != nil {
		return nil, err
	}
	curser.Next(ctx)
	if curser.Current == nil {
		return nil, ErrTxNotFound
	}
	rpcTx := &chainTypes.RpcTransaction{}
	err = curser.Decode(rpcTx)
	if err != nil {
		return nil, err
	}
	return rpcTx, nil
}

func (store *MongogDbStore) GetSendTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*chainTypes.RpcTransaction {
	rpcTx := []*chainTypes.RpcTransaction{}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	option := &options.FindOptions{}
	option.SetSkip(int64((pageIndex - 1) * pageSize))
	option.SetLimit(int64(pageSize))
	curser, err := store.txCol.Find(
		ctx,
		bson.M{"from": addr},
		option,
	)
	if err != nil {
		return rpcTx
	}
	err = curser.All(ctx, &rpcTx)
	if err != nil {
		return rpcTx
	}
	return rpcTx
}

func (store *MongogDbStore) GetReceiveTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*chainTypes.RpcTransaction {
	rpcTx := []*chainTypes.RpcTransaction{}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	option := &options.FindOptions{}
	option.SetSkip(int64((pageIndex - 1) * pageSize))
	option.SetLimit(int64(pageSize))
	curser, err := store.txCol.Find(
		ctx,
		bson.M{"to": addr},
		option,
	)
	if err != nil {
		return rpcTx
	}
	err = curser.All(ctx, &rpcTx)
	if err != nil {
		return rpcTx
	}
	return rpcTx
}

// Close disconnect db connection
// NOTICE Disconnect very slow, please wait
func (store *MongogDbStore) Close() {
	store.client.Disconnect(nil)
}
