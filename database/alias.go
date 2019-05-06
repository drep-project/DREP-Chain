package database

import (
	"fmt"
)

func constructAliasKey(inputKey []byte) string {
	return aliasKey + bytes2Hex(inputKey)
}

func (db *Database) AliasPut(key, value []byte) error {
	k := constructAliasKey(key)
	v := bytes2Hex(value)

	//在缓存中，地址是否已经被设置过别名
	if _, ok := db.aliasAddress.GetInverse(v); ok {
		return fmt.Errorf("in aliasAddress map ,addr has been set alias")
	}
	db.aliasAddress.Insert(k, v)
	return nil
}

func (db *Database) AliasGet(key []byte) ([]byte, error) {
	k := constructAliasKey(key)
	value, err := db.db.Get([]byte(k), nil)
	return value, err
}

func (db *Database) AliasDelete(key []byte) error {
	k := constructAliasKey(key)
	db.aliasAddress.Insert(k, nil)
	return nil
}

func (db *Database) aliasCommit() error {
	for k, v := range db.aliasAddress.GetForwardMap() {
		key := hex2Bytes(k.(string))
		if v != nil {
			err := db.put(key, []byte(hex2Bytes(v.(string))), false)
			if err != nil {
				return err
			}
		} else {
			err := db.delete(key, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
