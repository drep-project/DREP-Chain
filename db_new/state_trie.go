package db_new

import (
    "BlockChainTest/mycrypto"
    "errors"
)

var db *Database

type State struct {
    Sequence    string
    ChildrenKey [17][]byte
    Value       []byte
    IsLeaf      bool
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
        if state.ChildrenKey[i] != nil {
            children[i], _ = db.getState(state.ChildrenKey[i])
        }
    }
    return children
}

func (state *State) absorbChild(key []byte, child *State) {
    state.Sequence += child.Sequence
    state.Value = child.Value
    state.IsLeaf = child.IsLeaf
    for i := 0; i < 17; i++ {
        if child.ChildrenKey[i] != nil {
            st, err := db.getState(child.ChildrenKey[i])
            if err == nil {
                state.ChildrenKey[i] = getChildKey(key, i)
                db.putState(state.ChildrenKey[i], st)
                db.deleteState(child.ChildrenKey[i])
            }
        }
    }
    db.putState(key, state)
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
    children := state.getChildren()
    length, prefix := getCommonPrefix(seq, state.Sequence)
    i := getNextNibble(length, seq)
    j := getNextNibble(length, state.Sequence)
    if prefix == state.Sequence {
        if state.IsLeaf {
            if seq == state.Sequence {
                state.Value = value
            } else {
                state.ChildrenKey[16] = getChildKey(key, 16)
                children[16], err = newLeaf("", state.ChildrenKey[16], value)
                if err != nil {
                    return nil, err
                }
                state.ChildrenKey[i] = getChildKey(key, i)
                children[i], err = insert(seq[length:], state.ChildrenKey[i], value)
                if err != nil {
                    return nil, err
                }
                state.resetValue(children)
                state.IsLeaf = false
            }
        } else {
            state.ChildrenKey[i] = getChildKey(key, i)
            children[i], err = insert(seq[length:], state.ChildrenKey[i], value)
            if err != nil {
                return nil, err
            }
            state.resetValue(children)
        }
        err = db.putState(key, state)
        if err != nil {
            return nil, err
        }
        if state.ChildrenKey[i] != nil {
            err = db.putState(state.ChildrenKey[i], children[i])
            if err != nil {
                return nil, err
            }
        }
        return state, nil
    } else {
        state.Sequence = state.Sequence[length:]
        st := &State{}
        st.Sequence = prefix
        st.ChildrenKey[i] = getChildKey(key, i)
        st.ChildrenKey[j] = getChildKey(key, j)
        children[i], err = insert(seq[length:], st.ChildrenKey[i], value)
        if err != nil {
            return nil, err
        }
        children[j] = state
        st.resetValue(children)
        state = st
        db.putState(key, st)
        db.putState(st.ChildrenKey[i], children[i])
        db.putState(st.ChildrenKey[j], children[j])
        return state, nil
    }
}

func delete(key []byte, seq string) (*State, error) {
    state, err := db.getState(key)
    if err != nil {
        return nil, err
    }
    if state.IsLeaf && seq == state.Sequence {
        return state, nil
    }
    commonLen, _ := getCommonPrefix(seq, state.Sequence)
    if commonLen < len(state.Sequence) {
        return state, nil
    }
    children := state.getChildren()
    nib := getNextNibble(commonLen, seq)
    if state.ChildrenKey[nib] == nil {
        state.ChildrenKey[nib] = getChildKey(key, nib)
    }
    _, err = delete(state.ChildrenKey[nib], seq[commonLen:])
    if err != nil {
        return nil, err
    }
    children[nib] = nil
    if err == nil {
        sum := 0
        var uniqueChild *State
        for _, child := range children {
            if child != nil {
                sum += 1
                uniqueChild = child
            }
        }
        if sum == 1 {
            state.absorbChild(key, uniqueChild)
            return state, nil
        }
    }
    state.resetValue(children)
    db.putState(key, state)
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
        return nil, errors.New("node not found")
    }
    commonLen, _ := getCommonPrefix(seq, state.Sequence)
    if commonLen < len(state.Sequence) {
        return nil, errors.New("node not found")
    }
    nib := getNextNibble(commonLen, seq)
    return get(state.ChildrenKey[nib], seq[commonLen:])
}