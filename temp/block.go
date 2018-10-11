package temp

import (
	"BlockChainTest/bean"
	"BlockChainTest/database"
	"errors"
)

var (
	ErrWrongBlockKey = errors.New("get non-block type element")
)

func saveBlock(db *database.Database, block *bean.Block) error {
	_, _, err := db.Put(block)
	return err
}

func SaveBlock(block *bean.Block) error {
	db := database.GetDatabase()
	//fmt.Println("db: ", db)
	db.Open()
	defer db.Close()
	return saveBlock(db, block)
}

func SaveAllBlock(blocks []*bean.Block) error {
	db := database.GetDatabase()
	db.Open()
	defer db.Close()
	for i := 0; i < len(blocks); i ++ {
		saveBlock(db, blocks[i])
	}
	return nil
}

func loadBlock(db *database.Database, height int64) (*bean.Block, error) {
	key := bean.HeightToKey(height)
	elem, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	if block, ok := elem.(*bean.Block); ok {
		return block, nil
	} else {
		return nil, ErrWrongBlockKey
	}
}

func LoadBlock(height int64) (*bean.Block, error) {
	db := database.GetDatabase()
	db.Open()
	defer db.Close()
	return loadBlock(db, height)
}

func LoadAllBlock(fromHeight int64) []*bean.Block {
	db := database.GetDatabase()
	db.Open()
	defer db.Close()

	var (
		currentBlock *bean.Block
		err error
		height = fromHeight
		blocks = make([]*bean.Block, 0)
	)
	for err == nil {
		currentBlock, err = loadBlock(db, height)
		if err == nil {
			blocks = append(blocks, currentBlock)
		}
		height += 1
	}
	return blocks
}