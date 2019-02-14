package db_new

import (
    "encoding/hex"
    "BlockChainTest/mycrypto"
)

func bytes2Hex(key []byte) string {
    return hex.EncodeToString(key)
}

func hex2Bytes(key string) []byte {
    b, _ := hex.DecodeString(key)
    return b
}

func commonKey2TrieKey(key []byte) string {
    return bytes2Hex(mycrypto.Hash256(key, []byte("trie_key")))
}

func commonValue2TrieValue(value []byte) []byte {
    return mycrypto.Hash256(value, []byte("trie_value"))
}

func getPasserBy

func getChildKey(key []byte, child string) []byte {
    return mycrypto.Hash256(key, []byte("child"), []byte(child))
}

func getLeafBoolKey(key []byte) []byte {
    return mycrypto.Hash256(key, []byte("is leaf node"))
}