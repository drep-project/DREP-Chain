package db

import "BlockChainTest/config"

func (db *Database) GetChainStateRoot(chainId config.ChainIdType) []byte {
	if t, exists := db.tries[chainId]; exists {
		return t.Root.Value
	} else {
		return nil
	}
}