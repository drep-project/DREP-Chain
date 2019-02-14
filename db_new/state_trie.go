package db_new

import (
    "fmt"
    "BlockChainTest/mycrypto"
    "encoding/hex"
    "math"
)

var digits = [17]string{"", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}

func getTrieHexKey(key []byte) string {
    return hex.EncodeToString(key)
}

func GetTrieValue(value []byte) []byte {
    return mycrypto.Hash256(value)
}

type Node struct {
    mark     string
    children [17][]byte
    value    []byte
    leaf     bool
}







type StateNode struct {
    Key      string
    Value    []byte
    Children map[string] *StateNode
    IsLeaf   bool
}

func NewStateNode() *StateNode {
    n := &StateNode{}
    n.Children = make(map[string] *StateNode)
    return n
}

func (n *StateNode) resetValue() {
    hashList := make([][]byte, len(digits))
    for i, digit := range digits {
        if n.Children[digit] != nil {
            hashList[i] = n.Children[digit].Value
        }
    }
    ret := mycrypto.StackHash(hashList)
    n.Value = make([]byte, len(ret))
    copy(n.Value, ret)
    return
}

func (n *StateNode) setValue(value []byte) {
    n.Value = make([]byte, len(value))
    copy(n.Value, value)
}

func newLeaf(key string, value []byte) *StateNode {
    s := NewStateNode()
    s.Key = key
    s.setValue(value)
    s.IsLeaf = true
    return s
}

func insertNode(n *StateNode, key string, value []byte) *StateNode {
    if n == nil {
        return newLeaf(key, value)
    }
    commonLen, commonPrefix := getCommonPrefix(key, n.Key)
    dt0 := getNextDigit(commonLen, key)
    dt1 := getNextDigit(commonLen, n.Key)
    if commonPrefix == n.Key {
        if n.IsLeaf {
            if key == n.Key {
                n.setValue(value)
            } else {
                n.Children[""] = newLeaf("", n.Value)
                n.Children[dt0] = insertNode(n.Children[dt0], key[commonLen:], value)
                n.resetValue()
                n.IsLeaf = false
            }
        } else {
            n.Children[dt0] = insertNode(n.Children[dt0], key[commonLen:], value)
            n.resetValue()
        }
        return n
    }
    n.Key = n.Key[commonLen:]
    node := NewStateNode()
    node.Key = commonPrefix
    node.Children[dt0] = insertNode(node.Children[dt0], key[commonLen:], value)
    node.Children[dt1] = n
    node.resetValue()
    n = node
    return n
}

func insertNode123(db *Database, cur, key, value []byte, mark string) ([]byte, string) {
    v, _ := db.get(cur, true)
    if v == nil {
        db.put(key, value, true)
        return nil, ""
    }
    curMarkKey := getMarkKey(cur)
    curMarkValue, _  := db.get(curMarkKey, true)
    curMark := bytes2Hex(curMarkValue)
    commonLen, commonPrefix := getCommonPrefix(mark, curMark)
    dt0 := getNextDigit(commonLen, mark)
    dt1 := getNextDigit(commonLen, curMark)
    if commonPrefix == curMark {
        curLeafBoolValue, _ := db.get(getLeafBoolKey(cur), true)
        curLeafBool :=
        if n.IsLeaf {
            if key == n.Key {
                n.setValue(value)
            } else {
                n.Children[""] = newLeaf("", n.Value)
                n.Children[dt0] = insertNode(n.Children[dt0], key[commonLen:], value)
                n.resetValue()
                n.IsLeaf = false
            }
        } else {
            n.Children[dt0] = insertNode(n.Children[dt0], key[commonLen:], value)
            n.resetValue()
        }
        return n
    }

}

func deleteNode(n *StateNode, key string) (*StateNode, bool) {
    if n == nil {
        return nil, false
    }
    if n.IsLeaf && key == n.Key {
        return n, true
    }
    commonLen, _ := getCommonPrefix(key, n.Key)
    if commonLen < len(n.Key) {
        return n, false
    }
    dt := getNextDigit(commonLen, key)
    node, ok := deleteNode(n.Children[dt], key[commonLen:])
    if ok {
        sum := 0
        var uniqueChild *StateNode
        for str, child := range n.Children {
            if child == node {
                n.Children[str] = nil
                continue
            }
            if child != nil {
                sum += 1
                uniqueChild = child
            }
        }
        if sum == 1 {
            n.Key += uniqueChild.Key
            n.setValue(uniqueChild.Value)
            n.Children = uniqueChild.Children
            n.IsLeaf = uniqueChild.IsLeaf
            return n, false
        }
    }
    n.resetValue()
    return n, false
}

func getNode(n *StateNode, key string) *StateNode {
    if n == nil {
        return nil
    }
    if n.IsLeaf {
        if key == n.Key {
            return n
        }
        return nil
    }
    commonLen, _ := getCommonPrefix(key, n.Key)
    if commonLen < len(n.Key) {
        return nil
    }
    dt := getNextDigit(commonLen, key)
    return getNode(n.Children[dt], key[commonLen:])
}

func searchNode(n *StateNode, depth int) {
    fmt.Println()
    fmt.Println("depth: ", depth)
    fmt.Println("node: ", n)
    fmt.Println()
    for _, child := range n.Children {
        if child != nil {
            searchNode(child, depth+ 1)
        }
    }
}

type StateTrie struct {
    Root *StateNode
}

func NewStateTrie() *StateTrie {
    return &StateTrie{}
}

func (t *StateTrie) Insert(key, value []byte) {
    t.Root = insertNode(t.Root, getTrieHexKey(key), GetTrieValue(value))
}

func (t *StateTrie) Delete(key []byte) {
    t.Root, _ = deleteNode(t.Root, getTrieHexKey(key))
}

func (t *StateTrie) Get(key []byte) []byte {
    n := getNode(t.Root, getTrieHexKey(key))
    if n == nil {
        return nil
    }
    return n.Value
}

func (t *StateTrie) Search() {
    searchNode(t.Root, 0)
}

func (t *StateTrie) Validate() {
    root := t.Root
    v0 := make([]byte, len(root.Value))
    copy(v0, root.Value)
    root.resetValue()
}

func GetMerkleRoot(ts []*StateTrie) []byte {
    if ts == nil || len(ts) == 0 {
        return nil
    }
    l := len(ts)
    height := int(math.Ceil(math.Log(float64(l)) / math.Log(2))) + 1
    hashes := make([][]byte, l)
    for i, t := range ts {
        hashes[i] = t.Root.Value
    }
    for i := 0; i < height - 1; i++ {
        n := len(hashes)
        m := n / 2
        if m * 2 < n {
            m++
        }
        temp := make([][]byte, m)
        for j := 0; j < m; j++ {
            if 2 * j + 1 < n {
                temp[j] = mycrypto.Hash256(hashes[2 * j], hashes[2 * j + 1])
            } else {
                temp[j] = mycrypto.Hash256(hashes[2 * j], hashes[2 * j])
            }
        }
        hashes = make([][]byte, m)
        copy(hashes, temp)
    }
    return hashes[0]
}