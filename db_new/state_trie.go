package db_new

import (
    "BlockChainTest/mycrypto"
    "errors"
    "fmt"
    "strconv"
)

var db *Database

type State struct {
    Sequence  string
    ChildKeys [17][]byte
    Value     []byte
    IsLeaf    bool
}

func (state *State) resetValue(children [17]*State) {
    stack := make([]byte, 32 * 17)
    for i := 0; i < 17; i++ {
        if children[i] != nil && children[i].Value != nil {
            copy(stack[i * 17: (i + 1) * 17], children[i].Value)
        }
    }
    state.Value = mycrypto.Hash256(stack)
}

func (state *State) getChildren() [17]*State {
    var children [17]*State
    for i := 0; i < 17; i++ {
        if state.ChildKeys[i] != nil {
            children[i], _ = db.getState(state.ChildKeys[i])
        }
    }
    return children
}

func getChildKey(state *State, key []byte, nib int) []byte {
    if state.ChildKeys[nib] != nil {
        return state.ChildKeys[nib]
    }
    return mycrypto.Hash256(key, []byte("child"), []byte(strconv.Itoa(nib)))
}

func newLeaf(seq string, key, value []byte) (*State, error) {
    leaf := &State{
        Sequence: seq,
        Value:    value,
        IsLeaf:   true,
    }
    err := db.putState(key, leaf)
    if err != nil {
        return nil, err
    }
    return leaf, nil
}

func insert(seq string, key, value []byte) (*State, error) {
    state, err := db.getState(key)
    if err != nil {
        return newLeaf(seq, key, value)
    }
    prefix, offset := getCommonPrefix(seq, state.Sequence)
    if prefix == state.Sequence {
        if state.IsLeaf {
            if seq == state.Sequence {
                return insertExistedLeafValue(state, key, value)
            } else {
                return insertNewChildBranchOnLeaf(state, seq, offset, key, value)
            }
        } else {
            return insertProceedingOnCurrentBranch(state, seq, offset, key, value)
        }
    } else {
        return insertDivergingBranch(state, prefix, seq, offset, key, value)
    }
}

func insertExistedLeafValue(state *State, key, value []byte) (*State, error) {
    fmt.Println("call insertExistedLeafValue")
    fmt.Println("key:   ", bytes2Hex(key))
    fmt.Println("value: ", value)
    fmt.Println()

    state.Value = value
    err := db.putState(key, state)
    if err != nil {
        return nil, err
    }
    return state, nil
}

func insertNewChildBranchOnLeaf(state *State, seq string, offset int, key, value []byte) (*State, error) {
    fmt.Println("call insertNewChildBranchOnLeaf")
    fmt.Println("key:   ", bytes2Hex(key))
    fmt.Println("value: ", value)
    fmt.Println("seq:   ", seq)
    fmt.Println()

    var err error
    children := state.getChildren()

    state.ChildKeys[16] = getChildKey(state, key, 16)
    children[16], err = newLeaf("", state.ChildKeys[16], state.Value)
    if err != nil {
        return nil, err
    }

    nib := getNextNibble(seq, offset)
    state.ChildKeys[nib] = getChildKey(state, key, nib)
    children[nib], err = newLeaf(seq[offset:], state.ChildKeys[nib], value)
    if err != nil {
        return nil, err
    }

    state.resetValue(children)
    state.IsLeaf = false
    err = db.putState(key, state)
    if err != nil {
        return nil, err
    }
    return state, nil
}

func insertProceedingOnCurrentBranch(state *State, seq string, offset int, key, value []byte) (*State, error) {
    fmt.Println("call insertProceedingOnCurrentBranch")
    fmt.Println("key:   ", bytes2Hex(key))
    fmt.Println("value: ", value)
    fmt.Println("seq:   ", seq)
    fmt.Println()

    var err error
    children := state.getChildren()

    nib := getNextNibble(seq, offset)
    state.ChildKeys[nib] = getChildKey(state, key, nib)
    children[nib], err = insert(seq[offset:], state.ChildKeys[nib], value)
    if err != nil {
        return nil, err
    }
    db.putState(state.ChildKeys[nib], children[nib])
    if err != nil {
        return nil, err
    }

    state.resetValue(children)
    err = db.putState(key, state)
    if err != nil {
        return nil, err
    }
    err = db.putState(key, state)
    if err != nil {
        return nil, err
    }
    return state, nil
}

func insertDivergingBranch(state *State, prefix, seq string, offset int, key, value []byte) (*State, error) {
    fmt.Println("call insertDivergingBranch")
    fmt.Println("key:   ", bytes2Hex(key))
    fmt.Println("value: ", value)
    fmt.Println("seq:   ", seq)
    fmt.Println()

    var err error
    children := state.getChildren()

    nib0 := getNextNibble(state.Sequence, offset)
    childKey0 := getChildKey(state, key, nib0)
    div := &State{}
    div.Sequence = state.Sequence[offset:]
    for i, child := range children {
        if state.ChildKeys[i] != nil {
            div.ChildKeys[i] = getChildKey(div, state.ChildKeys[nib0], i)
            err = db.putState(div.ChildKeys[i], child)
            if err != nil {
                return nil, err
            }
            err = db.delState(state.ChildKeys[i])
            if err != nil {
                return nil, err
            }
        }
    }
    div.IsLeaf = state.IsLeaf
    if div.IsLeaf {
        div.Value = state.Value
    } else {
        div.resetValue(children)
    }
    err = db.putState(childKey0, div)
    if err != nil {
        return nil, err
    }

    nib1 := getNextNibble(seq, offset)
    childKey1 := getChildKey(state, key, nib1)
    leaf, err := newLeaf(seq[offset:], childKey1, value)
    if err != nil {
        return nil, err
    }

    var twins [17]*State
    twins[nib0] = div
    twins[nib1] = leaf
    for i, _ := range state.ChildKeys {
        state.ChildKeys[i] = nil
    }
    state.ChildKeys[nib0] = childKey0
    state.ChildKeys[nib1] = childKey1
    state.resetValue(twins)
    state.Sequence = prefix
    state.IsLeaf = false
    err = db.putState(key, state)
    if err != nil {
        return nil, err
    }

    return state, nil
}

func del(key []byte, seq string) (*State, error) {
    state, err := db.getState(key)
    if err != nil {
        return nil, err
    }
    if state.IsLeaf {
        if seq == state.Sequence {
            return delExistedLeaf(state, key)
        } else {
            return nil, errors.New("current key not found")
        }
    }
    return delProceedingOnCurrentBranch(state, seq, key)
}

func delExistedLeaf(state *State, key []byte) (*State, error) {
    err := db.delState(key)
    if err != nil {
        return nil, err
    }
    return state, nil
}

func delProceedingOnCurrentBranch(state *State, seq string, key []byte) (*State, error) {
    _, offset := getCommonPrefix(seq, state.Sequence)
    if offset < len(state.Sequence) {
        return nil, errors.New("current key not found")
    }
    nib := getNextNibble(seq, offset)
    state.ChildKeys[nib] = getChildKey(state, key, nib)
    _, err := del(state.ChildKeys[nib], seq[offset:])
    if err != nil {
        return nil, err
    }
    childCount, onlyChild := countChildren(state)
    if childCount == 1 {
        return absorbOnlyChild(state, onlyChild, key)
    }
    return state, nil
}

func countChildren(state *State) (int, *State) {
    children := state.getChildren()
    childCount := 0
    var onlyChild *State
    for i, childKey := range state.ChildKeys {
        if childKey != nil && children[i] != nil {
            childCount += 1
            onlyChild = children[i]
        }
    }
    return childCount, onlyChild
}

func absorbOnlyChild(state, onlyChild *State, key []byte) (*State, error) {
    fmt.Println("call absorbOnlyChild")
    state.Sequence += onlyChild.Sequence
    state.Value = onlyChild.Value
    state.IsLeaf = onlyChild.IsLeaf
    state.ChildKeys = onlyChild.ChildKeys
    err := db.putState(key, state)
    if err != nil {
        return nil, err
    }
    return state, nil
}

func get(key []byte, seq string) (*State, error) {
    state, err := db.getState(key)
    if err != nil {
        return nil, err
    }
    if state.IsLeaf {
        if seq == state.Sequence {
            return state, nil
        }
        return nil, errors.New("current key not found")
    }
    return getProceedingOnCurrentBranch(state, seq)
}

func getProceedingOnCurrentBranch(state *State, seq string) (*State, error) {
    _, offset := getCommonPrefix(seq, state.Sequence)
    if offset < len(state.Sequence) {
        return nil, errors.New("current key not found")
    }
    nib := getNextNibble(seq, offset)
    return get(state.ChildKeys[nib], seq[offset:])
}

func search(key []byte, seq string, depth int) {

    fmt.Println("current depth: ", depth)

    state, _ := db.getState(key)
    seq0 := seq
    seq += state.Sequence

    fmt.Println("seq old: ", seq0)
    fmt.Println("seq cat: ", state.Sequence)
    fmt.Println("seq:     ", seq)
    fmt.Println("value:   ", state.Value)
    fmt.Println("leaf:    ", state.IsLeaf)
    fmt.Println()

    for i := 0; i < 17; i++ {
        if state.ChildKeys[i] != nil {
            search(state.ChildKeys[i], seq, depth + 1)
        }
    }
}