package database

import (
	"BlockChainTest/bean"
	"errors"
)

var (
	ErrWrongBlockKey = errors.New("get non-block type element")
)

func saveBlock(db *Database, block *bean.Block) error {
	_, _, err := db.Put(block)
	return err
}

func SaveBlock(block *bean.Block) error {
	db := GetDatabase()
	//fmt.Println("db: ", db)
	//db.Open()
	//defer db.Close()
	return saveBlock(db, block)
}

func SaveAllBlock(blocks []*bean.Block) error {
	db := GetDatabase()
	//db.Open()
	//defer db.Close()
	for i := 0; i < len(blocks); i ++ {
		saveBlock(db, blocks[i])
	}
	return nil
}

func loadBlock(db *Database, height int64) (*bean.Block, error) {
	key := bean.Height2Key(height)
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
	db := GetDatabase()
	//db.Open()
	//defer db.Close()
	return loadBlock(db, height)
}

func LoadAllBlock(fromHeight int64) []*bean.Block {
	db := GetDatabase()
	//db.Open()
	//defer db.Close()

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

func PutInt(key string, value int) {
	db := GetDatabase()
	db.PutInt(key, value)
}

func GetInt(key string) (int, error) {
	db := GetDatabase()
	//db.Open()
	//defer db.Close()
	return db.GetInt(key)
}