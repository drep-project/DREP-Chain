package trie

import (
    "fmt"
    "BlockChainTest/crypto"
    "bytes"
)

var digits = [17]string{"", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}

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
    ret := crypto.StackHash(hashList)
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

func (t *StateTrie) Insert(key string, value []byte) {
    t.Root = insertNode(t.Root, key, value)
}

func (t *StateTrie) Delete(key string) {
    t.Root, _ = deleteNode(t.Root, key)
}

func (t *StateTrie) Get(key string) []byte {
    n := getNode(t.Root, key)
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
    fmt.Println()
    fmt.Println("result: ", bytes.Equal(v0, root.Value))
}