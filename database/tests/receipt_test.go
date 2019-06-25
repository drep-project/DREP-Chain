package tests

import (
    "testing"
    "github.com/drep-project/drep-chain/chain/service/chainservice"
    "fmt"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/database"
    "math/big"
    "github.com/drep-project/drep-chain/common"
    "time"
    "github.com/drep-project/drep-chain/app"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
)

var (
    db        *database.Database
    dbService *database.DatabaseService
    chain *chainservice.ChainService
    start uint64 = 0
    end uint64 = 2
)

func init() {
    db, _ = database.NewDatabase("test_db")
    dbService = database.NewDatabaseService(db)
    chain = &chainservice.ChainService{DatabaseService: dbService}
}

func TestPutAndGetReceipt(t *testing.T) {
    addr0 := crypto.Bytes2Address([]byte{1, 2, 3, 4})
    addr1 := crypto.Bytes2Address([]byte{5, 6, 7, 8})
    addr2 := crypto.Bytes2Address([]byte{10, 11, 12, 13})
    addr3 := crypto.Bytes2Address([]byte{17, 18, 19, 20})
    addrs := []crypto.CommonAddress{addr0, addr1, addr2, addr3}
    nonces := []uint64{0, 0, 0, 0}
    bal := new(big.Int).SetInt64(100000000000)
    db.PutBalance(&addr0, bal)
    db.PutBalance(&addr1, bal)
    db.PutBalance(&addr2, bal)
    db.PutBalance(&addr3, bal)
    fmt.Println("bal0: ", db.GetBalance(&addr0))
    fmt.Println("bal1: ", db.GetBalance(&addr1))
    fmt.Println("bal2: ", db.GetBalance(&addr2))
    fmt.Println("bal3: ", db.GetBalance(&addr3))
    var height, i uint64
    amount := new(common.Big)
    amount.SetMathBig(*new(big.Int).SetInt64(5))
    gasPrice := new(common.Big)
    gasPrice.SetMathBig(*new(big.Int).SetInt64(1))
    gasLimit := new(common.Big)
    gasLimit.SetMathBig(*new(big.Int).SetInt64(1000000000))
    blocks := make([]*types.Block, end - start)
    for height = start; height < end; height++ {
        txs := make([]*types.Transaction, 2)
        for i = 0; i < 2; i++ {
            fromIndex := (int) (height + i) % 4
            toIndex := (int) (height + i + 1) % 4
            tx := &types.Transaction{
                Data: types.TransactionData{
                    Version: common.Version,
                    Nonce: nonces[fromIndex],
                    Type: types.TransferType,
                    To: addrs[toIndex],
                    Amount: *amount,
                    GasPrice: *gasPrice,
                    GasLimit: *gasLimit,
                    Timestamp: time.Now().Unix(),
                    Data: addrs[fromIndex].Bytes(),
                },
            }
            txs[i] = tx
            nonces[fromIndex] += 1
        }
        blocks[height] = &types.Block{
            Header: &types.BlockHeader{
                ChainId: app.ChainIdType{},
                Version:     common.Version,
                PreviousHash: crypto.Hash{},
                GasLimit:    *gasLimit.ToInt(),
                GasUsed: *gasLimit.ToInt(),
                Height:      height,
                Timestamp:   uint64(time.Now().Unix()),
                StateRoot: []byte{0},
                TxRoot: []byte{0},
                LeaderPubKey: *secp256k1.NewPublicKey(new(big.Int).SetInt64(1), new(big.Int).SetInt64(2)),
                MinorPubKeys: []secp256k1.PublicKey{},
            },
            Data: &types.BlockData{
                TxCount: 2,
                TxList: txs,
            },
        }
        db.PutBlock(blocks[height])
    }
    stateProcessor := chainservice.NewStateProcessor(chain)
    //txValidator := chainservice.NewTransactionValidator(chain)
    for _, block := range blocks {
        receipts := make([]*types.Receipt, block.Data.TxCount)
        gp := new(chainservice.GasPool).AddGas(block.Header.GasLimit.Uint64() * 100)
        var usedGas uint64 = 0
        for i, tx := range block.Data.TxList {
            //receipt, _, _, _  := txValidator.ExecuteTransaction(db, tx, gp, block.Header)
            fmt.Println("data data: ", tx.Data.Data)
            from := crypto.Bytes2Address(tx.Data.Data)
            receipt, _, _ := stateProcessor.ApplyTransaction(db, chain, gp, block.Header, tx, &from, &usedGas)
            receipts[i] = receipt
            db.PutReceipt(*tx.TxHash(), receipt)
            fmt.Println("tx key: ", tx.TxHash().String())
            fmt.Println("value: ", db.GetReceipt(*tx.TxHash()))
        }
        db.PutReceipts(*block.Header.Hash(), receipts)
        fmt.Println("bk key: ", block.Header.Hash().String())
        fmt.Println("value: ", db.GetReceipts(*block.Header.Hash()))
    }
}

func TestGetReceiptByBlock(t *testing.T) {
    //var height uint64
    //for height = start; height < end; height++ {
    //    header, _ := chain.GetBlockHeaderByHeight(height)
    //    hash := header.Hash()
    //    receipts := chain.DatabaseService.GetReceipts(*hash)
    //    fmt.Println("receipts length: ", len(receipts))
    //    for i, r := range receipts {
    //        fmt.Println("i: ", i)
    //        printInfo(r)
    //    }
    //}
}

func printInfo(r *types.Receipt) {
    fmt.Println("tx hash             : ", r.TxHash)
    fmt.Println("ret                 : ", r.Ret)
    fmt.Println("gas used            : ", r.GasUsed)
    fmt.Println("contract address    : ", r.ContractAddress)
    fmt.Println("cumulative gas used : ", r.CumulativeGasUsed)
    fmt.Println("gas fee             : ", r.GasFee)
    fmt.Println("post state          : ", r.PostState)
    fmt.Println("status              : ", r.Status)
    fmt.Println("logs                : ", r.Logs)
    if r.Logs != nil && len(r.Logs) > 0 {
        for j, log := range r.Logs {
            fmt.Println("log ", j)
            fmt.Println("tx hash  : ", log.TxHash)
            fmt.Println("height   : ", log.Height)
            fmt.Println("data     : ", log.Data)
            fmt.Println("chain id : ", log.ChainId)
            fmt.Println("address  : ", log.Address)
            fmt.Println("removed  : ", log.Removed)
            fmt.Println("topics   : ", log.Topics)
            if log.Topics != nil && len(log.Topics) > 0 {
                for k, h := range log.Topics {
                    fmt.Println("topic ", k)
                    fmt.Println(h)
                }
            }
        }
    }
}