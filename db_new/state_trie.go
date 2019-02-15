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

func (state *State) absorbChild(key []byte, child *State) error {
    state.Sequence += child.Sequence
    state.Value = child.Value
    state.IsLeaf = child.IsLeaf
    for i := 0; i < 17; i++ {
        if child.ChildrenKey[i] != nil {
            st, err := db.getState(child.ChildrenKey[i])
            if err == nil {
                state.ChildrenKey[i] = getChildKey(key, i)
                err := db.putState(state.ChildrenKey[i], st)
                if err != nil {
                    return err
                }
                err = db.delState(child.ChildrenKey[i])
                if err != nil {
                    return err
                }
            }
        }
    }
    err := db.putState(key, state)
    if err != nil {
        return err
    }
    return nil
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

//func insertBackup(seq string, key, value []byte) (*State, error) {
//   state, err := db.getState(key)
//   if err != nil {
//       return newLeaf(seq, key, value)
//   }
//   children := state.getChildren()
//   length, prefix := getCommonPrefix(seq, state.Sequence)
//   nib0 := getNextNibble(length, seq)
//   nib1 := getNextNibble(length, state.Sequence)
//   if prefix == state.Sequence {
//       if state.IsLeaf {
//           if seq == state.Sequence {
//               state.Value = value
//           } else {
//               state.ChildrenKey[16] = getChildKey(key, 16)
//               children[16], err = newLeaf("", state.ChildrenKey[16], value)
//               if err != nil {
//                   return nil, err
//               }
//               state.ChildrenKey[nib0] = getChildKey(key, nib0)
//               children[nib0], err = insertBackup(seq[length:], state.ChildrenKey[nib0], value)
//               if err != nil {
//                   return nil, err
//               }
//               state.resetValue(children)
//               state.IsLeaf = false
//           }
//       } else {
//           state.ChildrenKey[nib0] = getChildKey(key, nib0)
//           children[nib0], err = insertBackup(seq[length:], state.ChildrenKey[nib0], value)
//           if err != nil {
//               return nil, err
//           }
//           state.resetValue(children)
//       }
//       err = db.putState(key, state)
//       if err != nil {
//           return nil, err
//       }
//       if state.ChildrenKey[nib0] != nil {
//           err = db.putState(state.ChildrenKey[nib0], children[nib0])
//           if err != nil {
//               return nil, err
//           }
//       }
//       return state, nil
//   } else {
//       state.Sequence = state.Sequence[length:]
//       st := &State{}
//       st.Sequence = prefix
//       st.ChildrenKey[nib0] = getChildKey(key, nib0)
//       st.ChildrenKey[nib1] = getChildKey(key, nib1)
//       children[nib0], err = insertBackup(seq[length:], st.ChildrenKey[nib0], value)
//       if err != nil {
//           return nil, err
//       }
//       children[nib1] = state
//       st.resetValue(children)
//       state = st
//       err = db.putState(key, st)
//       if err != nil {
//           return nil, err
//       }
//       err = db.putState(st.ChildrenKey[nib0], children[nib0])
//       if err != nil {
//           return nil, err
//       }
//       err = db.putState(st.ChildrenKey[nib1], children[nib1])
//       if err != nil {
//           return nil, err
//       }
//       return state, nil
//   }
//}

func insert(seq string, key, value []byte) (*State, error) {
    state, err := db.getState(key)
    if err != nil {
        return newLeaf(seq, key, value)
    }
    children := state.getChildren()
    length, prefix := getCommonPrefix(seq, state.Sequence)
    nib0 := getNextNibble(length, seq)
    nib1 := getNextNibble(length, state.Sequence)
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
                state.ChildrenKey[nib0] = getChildKey(key, nib0)
                children[nib0], err = insert(seq[length:], state.ChildrenKey[nib0], value)
                if err != nil {
                    return nil, err
                }
                state.resetValue(children)
                state.IsLeaf = false
            }
        } else {
            state.ChildrenKey[nib0] = getChildKey(key, nib0)
            children[nib0], err = insert(seq[length:], state.ChildrenKey[nib0], value)
            if err != nil {
                return nil, err
            }
            state.resetValue(children)
        }
        err = db.putState(key, state)
        if err != nil {
            return nil, err
        }
        if state.ChildrenKey[nib0] != nil {
            err = db.putState(state.ChildrenKey[nib0], children[nib0])
            if err != nil {
                return nil, err
            }
        }
        return state, nil
    } else {
        state.Sequence = state.Sequence[length:]
        st := &State{}
        st.Sequence = prefix
        st.ChildrenKey[nib0] = getChildKey(key, nib0)
        st.ChildrenKey[nib1] = getChildKey(key, nib1)
        children[nib0], err = insert(seq[length:], st.ChildrenKey[nib0], value)
        if err != nil {
            return nil, err
        }
        children[nib1] = state
        st.resetValue(children)
        state = st
        err = db.putState(key, st)
        if err != nil {
            return nil, err
        }
        err = db.putState(st.ChildrenKey[nib0], children[nib0])
        if err != nil {
            return nil, err
        }
        err = db.putState(st.ChildrenKey[nib1], children[nib1])
        if err != nil {
            return nil, err
        }
        return state, nil
    }
}

func del(key []byte, seq string) (*State, error) {
    state, err := db.getState(key)
    if err != nil {
        return nil, err
    }
    if state.IsLeaf && seq == state.Sequence {
        err = db.delState(key)
        if err != nil {
            return nil, err
        }
        return state, nil
    }
    commonLen, _ := getCommonPrefix(seq, state.Sequence)
    if commonLen < len(state.Sequence) {
        return state, nil
    }
    children := state.getChildren()
    nib := getNextNibble(commonLen, seq)
    state.ChildrenKey[nib] = getChildKey(key, nib)
    _, err = del(state.ChildrenKey[nib], seq[commonLen:])
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
            err = state.absorbChild(key, uniqueChild)
            if err != nil {
                return nil, err
            }
            return state, nil
        }
    }
    state.resetValue(children)
    err = db.putState(key, state)
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
        return nil, errors.New("node not found")
    }
    commonLen, _ := getCommonPrefix(seq, state.Sequence)
    if commonLen < len(state.Sequence) {
        return nil, errors.New("node not found")
    }
    nib := getNextNibble(commonLen, seq)
    return get(state.ChildrenKey[nib], seq[commonLen:])
}

func updateLeafValue(state *State, key, value []byte) (*State, error) {
    var err error
    state.Value = value
    err = db.putState(key, state)
    if err != nil {
        return nil, err
    }
    return state, nil
}

func addNewBranchOnLeaf(state *State, length int, seq string, key, value []byte) (*State, error) {
    var err error
    children := state.getChildren()

    state.ChildrenKey[16] = getChildKey(key, 16)
    children[16], err = newLeaf("", state.ChildrenKey[16], value)
    if err != nil {
        return nil, err
    }

    nib := getNextNibble(length, seq)
    state.ChildrenKey[nib] = getChildKey(key, nib)
    children[nib], err = insert(seq[length:], state.ChildrenKey[nib], value)
    if err != nil {
        return nil, err
    }

    state.resetValue(children)
    state.IsLeaf = false
    err = db.putState(key, state)
    if err != nil {
        return nil, err
    }
    err = db.putState(state.ChildrenKey[nib], children[nib])
    if err != nil {
        return nil, err
    }
    return state, nil
}

func proceedOnCurrentBranch(state *State, length int, seq string, key, value []byte) (*State, error) {
    var err error
    children := state.getChildren()

    nib := getNextNibble(length, seq)
    state.ChildrenKey[nib] = getChildKey(key, nib)
    children[nib], err = insert(seq[length:], state.ChildrenKey[nib], value)
    if err != nil {
        return nil, err
    }

    state.resetValue(children)
    err = db.putState(key, state)
    if err != nil {
        return nil, err
    }
    err = db.putState(state.ChildrenKey[nib], children[nib])
    if err != nil {
        return nil, err
    }
    return state, nil
}

func splitOnCurrentBranch(state *State, length int, prefix, seq string, key, value []byte) (*State, error) {
    var err error
    children := state.getChildren()

    nib0 := getNextNibble(length, state.Sequence)
    state.ChildrenKey[nib0] = getChildKey(key, nib0)
    div := &State{}
    div.Sequence = state.Sequence[length:]
    div.IsLeaf = state.IsLeaf
    for i, child := range children {
        if state.ChildrenKey[i] != nil {
            div.ChildrenKey[i] = getChildKey(state.ChildrenKey[nib0], i)
            err = db.putState(div.ChildrenKey[i], child)
            if err != nil {
                return nil, err
            }
            err = db.delState(state.ChildrenKey[i])
            if err != nil {
                return nil, err
            }
        }
    }
    div.resetValue(children)
    err = db.putState(state.ChildrenKey[nib0], div)
    if err != nil {
        return nil, err
    }

    nib1 := getNextNibble(length, seq)
    state.ChildrenKey[nib1] = getChildKey(key, nib1)
    leaf, err := newLeaf(seq[length:], state.ChildrenKey[nib1], value)
    if err != nil {
        return nil, err
    }
    err = db.putState(state.ChildrenKey[nib1], leaf)
    if err != nil {
        return nil, err
    }

    var twins [17]*State
    twins[nib0] = div
    twins[nib1] = leaf
    for i, _ := range state.ChildrenKey {
        if i != nib0 && i != nib1 {
            state.ChildrenKey[i] = nil
        }
    }
    state.resetValue(twins)
    state.Sequence = prefix
    state.IsLeaf = false
    err = db.putState(key, state)
    if err != nil {
        return nil, err
    }

    return state, nil
}