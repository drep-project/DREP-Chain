package trace

import (
	"bytes"
	"errors"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
	"time"
)

var (

)

func makeBlockPools() map[uint64]*types.Block{
	blockPool := make(map[uint64]*types.Block)
	n:=20
	blocks := seuqenceBlock(n)
	for i:=0;i<n;i++ {
		blockPool[uint64(i)] = blocks[i]
	}
	return blockPool
}



func Test_Atach_Block_Process(t *testing.T) {
	path := "./test_process_db1"
	config := HistoryConfig{path,"",  "leveldb",  true }
	blockPool := makeBlockPools()
	analysis := NewBlockAnalysis(config,  func (height uint64) (*types.Block, error) {return nil,nil})
	defer func() {
		analysis.Close()
		deleteFolder(path)
	}()
	var newBlockFeed event.Feed
	var detachBlockFeed event.Feed
	analysis.Start(&newBlockFeed, &detachBlockFeed)


	for _, block := range blockPool {
		newBlockFeed.Send(block)
	}

	time.Sleep(time.Second)
	for _, block := range blockPool {
		for _, tx := range block.Data.TxList {
			txBytes, err := analysis.store.GetRawTransaction(tx.TxHash())
			if err != nil {
				t.Error(err)
			}
			if !bytes.Equal(tx.AsPersistentMessage(), txBytes) {
				t.Errorf("tx raw data in db not match the actual tx data")
			}
		}
	}
}


func Test_Detach_Block_Process(t *testing.T) {
	path := "./test_process_db2"
	config := HistoryConfig{path,"",  "leveldb",  true }
	blockPool := makeBlockPools()
	analysis := NewBlockAnalysis(config,  func (height uint64) (*types.Block, error) {return nil,nil})
	defer func() {
		analysis.Close()
		deleteFolder(path)
	}()
	var newBlockFeed event.Feed
	var detachBlockFeed event.Feed
	analysis.Start(&newBlockFeed, &detachBlockFeed)

	for _, block := range blockPool {
		newBlockFeed.Send(block)
	}

	time.Sleep(time.Second)

	for _, block := range blockPool {
		detachBlockFeed.Send(block)
	}

	time.Sleep(time.Second)
	for _, block := range blockPool {
		for _, tx := range block.Data.TxList {
			_, err := analysis.store.GetRawTransaction(tx.TxHash())
			if err != leveldb.ErrNotFound {
				t.Errorf("expect the transaction to be deleted but the transaction still exists")
			}
		}
	}
}

func Test_Rebuild(t *testing.T) {
	path := "./test_process_db3"
	config := HistoryConfig{path,"",  "leveldb",  true }
	blockPool := makeBlockPools()
	analysis := NewBlockAnalysis(config,  func (height uint64) (*types.Block, error) {
		block, ok := blockPool[height]
		if ok {
			return block, nil
		}else{
			return nil, errors.New("block not found")
		}
	})
	defer func() {
		analysis.Close()
		deleteFolder(path)
	}()
	var newBlockFeed event.Feed
	var detachBlockFeed event.Feed
	analysis.Start(&newBlockFeed, &detachBlockFeed)

	removeArr := []int{3,6,7,8,12,19}
	for index, block := range blockPool {
		if !contain(removeArr, int(index)) {
			newBlockFeed.Send(block)
		}
	}
	time.Sleep(time.Second)
	analysis.Rebuild(0,20)
	for _, block := range blockPool {
		for _, tx := range block.Data.TxList {
			txBytes, err := analysis.store.GetRawTransaction(tx.TxHash())
			if err != nil {
				t.Error(err)
			}
			if !bytes.Equal(tx.AsPersistentMessage(), txBytes) {
				t.Errorf("tx raw data in db not match the actual tx data")
			}
		}
	}
}

func TestMain(m *testing.M) {
	m.Run()
}

func contain(arr []int,val int) bool {
	for i:= 0;i<len(arr);i++{
		if arr[i] == val {
			return true
		}
	}
	return false
}