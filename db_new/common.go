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

func getMarkKey(key []byte) []byte {
    return mycrypto.Hash256(key, []byte("mark"))
}

func getChildKey(key []byte, child string) []byte {
    return mycrypto.Hash256(key, []byte("child"), []byte(child))
}

func getLeafBoolKey(key []byte) []byte {
    return mycrypto.Hash256(key, []byte("is leaf node"))
}

func getCommonPrefix(s1, s2 string) (int, string) {
    if s1 == "" || s2 == "" {
        return 0, ""
    }
    for i := 0; i < len(s1); i++ {
        if i == len(s2) {
            return i, s2
        }
        if s1[i] == s2[i] {
            continue
        }
        return i, s1[:i]
    }
    return len(s1), s1
}

func getNextDigit(start int, str string) string {
    if start == len(str) {
        return str[start:]
    }
    return str[start: start + 1]
}