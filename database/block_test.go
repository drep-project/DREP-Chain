package database

import (
	"testing"
	"BlockChainTest/bean"
	"time"
	"fmt"
)

func demoBlock(height int64) *bean.Block {
	header := &bean.BlockHeader{Height: height, Timestamp: time.Now().UnixNano()}
	block := &bean.Block{Header: header}
	return block
}

func TestSave(t *testing.T) {
	blocks := make([]*bean.Block, 5)
	for i := 0; i < 5; i ++ {
		blocks[i] = demoBlock(int64(i + 1))
		time.Sleep(time.Second)
	}
	block := demoBlock(100)
	SaveBlock(block)
	SaveAllBlock(blocks)
}

// TODO change all tab indents into spaces
func TestLoad(t *testing.T) {
	block, err := LoadBlock(100)
	fmt.Println()
	fmt.Println("block: ", block)
	fmt.Println("err: ", err)
	fmt.Println()

	blocks := LoadAllBlock(1)
	fmt.Println()
	for i := 0; i < len(blocks); i ++ {
		fmt.Println(blocks[i])
	}
	fmt.Println()
}
