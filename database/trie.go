package database

import (
    "errors"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "strconv"
)

type Node struct {
    Sequence  string
    ChildKeys [17][]byte
    Value     []byte
    IsLeaf    bool
}

func (node *Node) resetValue(children [17]*Node) {
    stack := make([]byte, 32 * 17)
    for i := 0; i < 17; i++ {
        if children[i] != nil && children[i].Value != nil {
            copy(stack[i * 17: (i + 1) * 17], children[i].Value)
        }
    }
    node.Value = sha3.Hash256(stack)
}

func (node *Node) getChildKey(key []byte, nib int) []byte {
    if node.ChildKeys[nib] != nil {
        return node.ChildKeys[nib]
    }
    return sha3.HashS256(key, []byte("child"), []byte(strconv.Itoa(nib)))
}

type Trie interface {
    RootKey()                       []byte
    GetNode(key []byte)             (*Node, error)
    PutNode(key []byte, node *Node) error
    DelNode(key []byte)             error
}

func updateTrie(trie Trie, key, value []byte) error {
    seq := bytes2Hex(key)
    _, err := insert(trie, seq, trie.RootKey(), value)
    return err
}

func getRootValue(trie Trie) []byte {
    root, err := trie.GetNode(trie.RootKey())
    if err != nil {
        return nil
    }
    return root.Value
}

func getChildren(trie Trie, node *Node) [17]*Node {
    var children [17]*Node
    for i := 0; i < 17; i++ {
        if node.ChildKeys[i] != nil {
            children[i], _ = trie.GetNode(node.ChildKeys[i])
        }
    }
    return children
}

func countChildren(trie Trie, node *Node) (int, *Node) {
    children := getChildren(trie, node)
    childCount := 0
    var onlyChild *Node
    for i, childKey := range node.ChildKeys {
        if childKey != nil && children[i] != nil {
            childCount += 1
            onlyChild = children[i]
        }
    }
    return childCount, onlyChild
}

func newLeaf(trie Trie, seq string, key, value []byte) (*Node, error) {
    leaf := &Node{
        Sequence: seq,
        Value:    value,
        IsLeaf:   true,
    }
    err := trie.PutNode(key, leaf)
    if err != nil {
        return nil, err
    }
    return leaf, nil
}

func insert(trie Trie, seq string, key, value []byte) (*Node, error) {
    node, err := trie.GetNode(key)
    if err != nil {
        return newLeaf(trie, seq, key, value)
    }
    prefix, offset := getCommonPrefix(seq, node.Sequence)
    if prefix == node.Sequence {
        if node.IsLeaf {
            if seq == node.Sequence {
                return insertExistedLeafValue(trie, node, key, value)
            } else {
                return insertNewChildBranchOnLeaf(trie, node, seq, offset, key, value)
            }
        } else {
            return insertProceedingOnCurrentBranch(trie, node, seq, offset, key, value)
        }
    } else {
        return insertDivergingBranch(trie, node, prefix, seq, offset, key, value)
    }
}

func insertExistedLeafValue(trie Trie, node *Node, key, value []byte) (*Node, error) {
    node.Value = value
    err := trie.PutNode(key, node)
    if err != nil {
        return nil, err
    }
    return node, nil
}

func insertNewChildBranchOnLeaf(trie Trie, node *Node, seq string, offset int, key, value []byte) (*Node, error) {
    var err error
    children := getChildren(trie, node)

    node.ChildKeys[16] = node.getChildKey(key, 16)
    children[16], err = newLeaf(trie, "", node.ChildKeys[16], node.Value)
    if err != nil {
        return nil, err
    }

    nib := getNextNibble(seq, offset)
    node.ChildKeys[nib] = node.getChildKey(key, nib)
    children[nib], err = newLeaf(trie, seq[offset:], node.ChildKeys[nib], value)
    if err != nil {
        return nil, err
    }

    node.resetValue(children)
    node.IsLeaf = false
    err = trie.PutNode(key, node)
    if err != nil {
        return nil, err
    }
    return node, nil
}

func insertProceedingOnCurrentBranch(trie Trie, node *Node, seq string, offset int, key, value []byte) (*Node, error) {
    var err error
    children := getChildren(trie, node)

    nib := getNextNibble(seq, offset)
    node.ChildKeys[nib] = node.getChildKey(key, nib)
    children[nib], err = insert(trie, seq[offset:], node.ChildKeys[nib], value)
    if err != nil {
        return nil, err
    }
    trie.PutNode(node.ChildKeys[nib], children[nib])
    if err != nil {
        return nil, err
    }

    node.resetValue(children)
    err = trie.PutNode(key, node)
    if err != nil {
        return nil, err
    }
    return node, nil
}

func insertDivergingBranch(trie Trie, node *Node, prefix, seq string, offset int, key, value []byte) (*Node, error) {
    var err error
    children := getChildren(trie, node)

    nib0 := getNextNibble(node.Sequence, offset)
    childKey0 := node.getChildKey(key, nib0)
    div := &Node{}
    div.Sequence = node.Sequence[offset:]
    for i, child := range children {
        if node.ChildKeys[i] != nil {
            div.ChildKeys[i] = div.getChildKey(node.ChildKeys[nib0], i)
            err = trie.PutNode(div.ChildKeys[i], child)
            if err != nil {
                return nil, err
            }
            trie.DelNode(node.ChildKeys[i])
        }
    }
    div.IsLeaf = node.IsLeaf
    if div.IsLeaf {
        div.Value = node.Value
    } else {
        div.resetValue(children)
    }
    err = trie.PutNode(childKey0, div)
    if err != nil {
        return nil, err
    }

    nib1 := getNextNibble(seq, offset)
    childKey1 := node.getChildKey(key, nib1)
    leaf, err := newLeaf(trie, seq[offset:], childKey1, value)
    if err != nil {
        return nil, err
    }

    var twins [17]*Node
    twins[nib0] = div
    twins[nib1] = leaf
    for i, _ := range node.ChildKeys {
        node.ChildKeys[i] = nil
    }
    node.ChildKeys[nib0] = childKey0
    node.ChildKeys[nib1] = childKey1
    node.resetValue(twins)
    node.Sequence = prefix
    node.IsLeaf = false
    err = trie.PutNode(key, node)
    if err != nil {
        return nil, err
    }

    return node, nil
}

func del(trie Trie, key []byte, seq string) (*Node, error) {
    node, err := trie.GetNode(key)
    if err != nil {
        return nil, err
    }
    if node.IsLeaf {
        if seq == node.Sequence {
            return delExistedLeaf(trie, node, key)
        } else {
            return nil, errors.New("current key not found")
        }
    }
    return delProceedingOnCurrentBranch(trie, node, seq, key)
}

func delExistedLeaf(trie Trie, node *Node, key []byte) (*Node, error) {
    trie.DelNode(key)
    return node, nil
}

func delProceedingOnCurrentBranch(trie Trie, node *Node, seq string, key []byte) (*Node, error) {
    _, offset := getCommonPrefix(seq, node.Sequence)
    if offset < len(node.Sequence) {
        return nil, errors.New("current key not found")
    }
    nib := getNextNibble(seq, offset)
    node.ChildKeys[nib] = node.getChildKey(key, nib)
    _, err := del(trie, node.ChildKeys[nib], seq[offset:])
    if err != nil {
        return nil, err
    }
    childCount, onlyChild := countChildren(trie, node)
    if childCount == 1 {
        return absorbOnlyChild(trie, node, onlyChild, key)
    }
    return node, nil
}

func absorbOnlyChild(trie Trie, node, onlyChild *Node, key []byte) (*Node, error) {
    node.Sequence += onlyChild.Sequence
    node.Value = onlyChild.Value
    node.IsLeaf = onlyChild.IsLeaf
    node.ChildKeys = onlyChild.ChildKeys
    err := trie.PutNode(key, node)
    if err != nil {
        return nil, err
    }
    return node, nil
}

func get(trie Trie, key []byte, seq string) (*Node, error) {
    node, err := trie.GetNode(key)
    if err != nil {
        return nil, err
    }
    if node.IsLeaf {
        if seq == node.Sequence {
            return node, nil
        }
        return nil, errors.New("current key not found")
    }
    return getProceedingOnCurrentBranch(trie, node, seq)
}

func getProceedingOnCurrentBranch(trie Trie, node *Node, seq string) (*Node, error) {
    _, offset := getCommonPrefix(seq, node.Sequence)
    if offset < len(node.Sequence) {
        return nil, errors.New("current key not found")
    }
    nib := getNextNibble(seq, offset)
    return get(trie, node.ChildKeys[nib], seq[offset:])
}