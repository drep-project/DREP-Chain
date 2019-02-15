package db_new

import (
    "encoding/hex"
    "BlockChainTest/mycrypto"
    "strconv"
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

func getChildKey(key []byte, nibbleID int) []byte {
    return mycrypto.Hash256(key, []byte("child"), []byte(strconv.Itoa(nibbleID)))
}

func getCommonPrefix(seq1, seq2 string) (int, string) {
    if seq1 == "" || seq2 == "" {
        return 0, ""
    }
    for i := 0; i < len(seq1); i++ {
        if i == len(seq2) {
            return i, seq2
        }
        if seq1[i] == seq2[i] {
            continue
        }
        return i, seq1[:i]
    }
    return len(seq1), seq1
}

func getNextNibble(start int, mark string) int {
    if start == len(mark) {
        return 16
    }
    return char2Nibble(mark[start: start + 1])
}

func char2Nibble(char string) int {
    switch char {
    case "0":
        return 0
    case "1":
        return 1
    case "2":
        return 2
    case "3":
        return 3
    case "4":
        return 4
    case "5":
        return 5
    case "6":
        return 6
    case "7":
        return 7
    case "8":
        return 8
    case "9":
        return 9
    case "a":
        return 10
    case "b":
        return 11
    case "c":
        return 12
    case "d":
        return 13
    case "e":
        return 14
    case "f":
        return 15
    default:
        return 16
    }
}