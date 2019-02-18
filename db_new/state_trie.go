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
    state := &State{
        Sequence: seq,
        Value:    value,
        IsLeaf:   true,
    }
    err := db.putState(key, state)
    if err != nil {
        return nil, err
    }
    return state, nil
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
    err = db.putState(state.ChildKeys[16], children[16])
    if err != nil {
        return nil, err
    }

    nib := getNextNibble(seq, offset)
    state.ChildKeys[nib] = getChildKey(state, key, nib)
    children[nib], err = newLeaf(seq[offset:], state.ChildKeys[nib], value)
    if err != nil {
        return nil, err
    }
    err = db.putState(state.ChildKeys[nib], children[nib])
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
    state.ChildKeys[nib0] = getChildKey(state, key, nib0)
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
    err = db.putState(state.ChildKeys[nib0], div)
    if err != nil {
        fmt.Println("111111111")
        return nil, err
    }
    fmt.Println("div:       ")
    fmt.Println("nib0:      ", nib0)
    fmt.Println("div seq:   ", div.Sequence)
    fmt.Println("div key:   ", bytes2Hex(state.ChildKeys[nib0]))
    fmt.Println("div value: ", div.Value)
    fmt.Println()

    nib1 := getNextNibble(seq, offset)
    state.ChildKeys[nib1] = getChildKey(state, key, nib1)
    leaf, err := newLeaf(seq[offset:], state.ChildKeys[nib1], value)
    if err != nil {
        fmt.Println("222222222")
        return nil, err
    }
    err = db.putState(state.ChildKeys[nib1], leaf)
    if err != nil {
        fmt.Println("33333333")
        return nil, err
    }
    fmt.Println("nib1:     ", nib1)
    fmt.Println("leaf key: ", bytes2Hex(state.ChildKeys[nib1]))

    var twins [17]*State
    twins[nib0] = div
    twins[nib1] = leaf
    for i, _ := range state.ChildKeys {
        if i != nib0 && i != nib1 {
            state.ChildKeys[i] = nil
        }
    }
    fmt.Println("leaf key: ", bytes2Hex(state.ChildKeys[nib1]))
    state.resetValue(twins)
    state.Sequence = prefix
    state.IsLeaf = false
    err = db.putState(key, state)
    if err != nil {
        fmt.Println("4444444444")
        return nil, err
    }

    fmt.Println("555555555")

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
    children := state.getChildren()
    state.ChildKeys[nib] = nil
    children[nib] = nil
    childCount, onlyChild := countChildren(state, children)
    if childCount == 0 {
        return discardIsolatedNode(state, key)
    }
    if childCount == 1 {
        return absorbOnlyChild(state, onlyChild, key)
    }
    return state, nil
}

func countChildren(state *State, children [17]*State) (int, *State) {
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

func discardIsolatedNode(state *State, key []byte) (*State, error) {
    err := db.delState(key)
    if err != nil {
        return nil, err
    }
    return state, nil
}

func absorbOnlyChild(state, onlyChild *State, key []byte) (*State, error) {
    state.Sequence += onlyChild.Sequence
    state.Value = onlyChild.Value
    state.IsLeaf = onlyChild.IsLeaf
    for i := 0; i < 17; i++ {
        if onlyChild.ChildKeys[i] != nil {
            st, err := db.getState(onlyChild.ChildKeys[i])
            if err == nil {
                state.ChildKeys[i] = getChildKey(state, key, i)
                err := db.putState(state.ChildKeys[i], st)
                if err != nil {
                    return nil, err
                }
                err = db.delState(onlyChild.ChildKeys[i])
                if err != nil {
                    return nil, err
                }
            }
        }
    }
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
    fmt.Println("key:     ", bytes2Hex(key))
    fmt.Println()
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
            fmt.Println("searching i: ", i)
        }
    }
}