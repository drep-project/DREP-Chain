package db_new

import (
    "BlockChainTest/mycrypto"
    "errors"
)

var db *Database

type State struct {
    Sequence string
    Children [17][]byte
    Value    []byte
    IsLeaf   bool
}

func (state *State) resetValue(childStates [17]*State) {
    stack := make([]byte, 32 * 17)
    for i := 0; i < 17; i++ {
        if childStates[i] != nil && childStates[i].Value != nil {
            copy(stack[i * 17: (i + 1) * 17], childStates[i].Value)
        }
    }
    state.Value = mycrypto.Hash256(stack)
}

func (state *State) getChildNodes() [17]*State {
    var childNodes [17]*State
    for i := 0; i < 17; i++ {
        if state.Children[i] != nil {
            childNodes[i], _ = db.getState(state.Children[i])
        }
    }
    return childNodes
}

func (state *State) absorbChild(key []byte, child *State) {
    state.Sequence += child.Sequence
    state.Value = child.Value
    state.IsLeaf = child.IsLeaf
    for i := 0; i < 17; i++ {
        if child.Children[i] != nil {
            n, err := db.getState(child.Children[i])
            if err == nil {
                state.Children[i] = getChildKey(key, i)
                db.putState(state.Children[i], n)
            }
        }
    }
    db.putState(key, state)
}

func newLeaf(seq string, key, value []byte) (*State, error) {
    node := &State{
        Sequence: seq,
        Value:    value,
        IsLeaf:   true,
    }
    err := db.putState(key, node)
    if err != nil {
        return nil, err
    }
    return node, nil
}

func insert(seq string, key, value []byte) (*State, error) {
    n, err := db.getState(key)
    if err != nil {
        return newLeaf(seq, key, value)
    }
    childNodes := n.getChildNodes()
    commonLen, commonPrefix := getCommonPrefix(seq, n.Sequence)
    i := getNextNibble(commonLen, seq)
    j := getNextNibble(commonLen, n.Sequence)
    if commonPrefix == n.Sequence {
        if n.IsLeaf {
            if seq == n.Sequence {
                n.Value = value
            } else {
                n.Children[16] = getChildKey(key, 16)
                childNodes[16], err = newLeaf("", n.Children[16], value)
                if err != nil {
                    return nil, err
                }
                n.Children[i] = getChildKey(key, i)
                childNodes[i], err = insert(seq[commonLen:], n.Children[i], value)
                if err != nil {
                    return nil, err
                }
                n.resetValue(childNodes)
                n.IsLeaf = false
            }
        } else {
            n.Children[i] = getChildKey(key, i)
            childNodes[i], err = insert(seq[commonLen:], n.Children[i], value)
            if err != nil {
                return nil, err
            }
            n.resetValue(childNodes)
        }
        return n, nil
    } else {
        n.Sequence = n.Sequence[commonLen:]
        node := &State{}
        node.Sequence = commonPrefix
        node.Children[i] = getChildKey(key, i)
        node.Children[j] = getChildKey(key, j)
        childNodes[i], err = insert(seq[commonLen:], node.Children[i], value)
        if err != nil {
            return nil, err
        }
        childNodes[j] = n
        node.resetValue(childNodes)
        n = node
        return n, nil
    }
}

func delete(nKey []byte, mark string) (*State, error) {
    n, err := db.getState(nKey)
    if err != nil {
        return nil, err
    }
    if n.IsLeaf && mark == n.Sequence {
        return n, nil
    }
    commonLen, _ := getCommonPrefix(mark, n.Sequence)
    if commonLen < len(n.Sequence) {
        return n, nil
    }
    childNodes := n.getChildNodes()
    id := getNextNibble(commonLen, mark)
    if n.Children[id] == nil {
        n.Children[id] = getChildKey(nKey, id)
    }
    _, err = delete(n.Children[id], mark[commonLen:])
    if err != nil {
        return nil, err
    }
    childNodes[id] = nil
    if err == nil {
        sum := 0
        var uniqueChild *State
        for _, child := range childNodes {
            if child != nil {
                sum += 1
                uniqueChild = child
            }
        }
        if sum == 1 {
            n.absorbChild(nKey, uniqueChild)
            return n, nil
        }
    }
    n.resetValue(childNodes)
    return n, nil
}

func get(nKey []byte, mark string) (*State, error) {
    n, err := db.getState(nKey)
    if err != nil {
        return nil, err
    }
    if n.IsLeaf {
        if mark == n.Sequence {
            return n, nil
        }
        return nil, errors.New("node not found")
    }
    commonLen, _ := getCommonPrefix(mark, n.Sequence)
    if commonLen < len(n.Sequence) {
        return nil, errors.New("node not found")
    }
    id := getNextNibble(commonLen, mark)
    return get(n.Children[id], mark[commonLen:])
}