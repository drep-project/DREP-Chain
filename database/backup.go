package database

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
//               state.ChildKeys[16] = getChildKey(key, 16)
//               children[16], err = newLeaf("", state.ChildKeys[16], value)
//               if err != nil {
//                   return nil, err
//               }
//               state.ChildKeys[nib0] = getChildKey(key, nib0)
//               children[nib0], err = insertBackup(seq[length:], state.ChildKeys[nib0], value)
//               if err != nil {
//                   return nil, err
//               }
//               state.resetValue(children)
//               state.IsLeaf = false
//           }
//       } else {
//           state.ChildKeys[nib0] = getChildKey(key, nib0)
//           children[nib0], err = insertBackup(seq[length:], state.ChildKeys[nib0], value)
//           if err != nil {
//               return nil, err
//           }
//           state.resetValue(children)
//       }
//       err = db.putState(key, state)
//       if err != nil {
//           return nil, err
//       }
//       if state.ChildKeys[nib0] != nil {
//           err = db.putState(state.ChildKeys[nib0], children[nib0])
//           if err != nil {
//               return nil, err
//           }
//       }
//       return state, nil
//   } else {
//       state.Sequence = state.Sequence[length:]
//       st := &State{}
//       st.Sequence = prefix
//       st.ChildKeys[nib0] = getChildKey(key, nib0)
//       st.ChildKeys[nib1] = getChildKey(key, nib1)
//       children[nib0], err = insertBackup(seq[length:], st.ChildKeys[nib0], value)
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
//       err = db.putState(st.ChildKeys[nib0], children[nib0])
//       if err != nil {
//           return nil, err
//       }
//       err = db.putState(st.ChildKeys[nib1], children[nib1])
//       if err != nil {
//           return nil, err
//       }
//       return state, nil
//   }
//}

//func delBackup(key []byte, seq string) (*State, error) {
//    state, err := db.getState(key)
//    if err != nil {
//        return nil, err
//    }
//    if state.IsLeaf && seq == state.Sequence {
//        err = db.delState(key)
//        if err != nil {
//            return nil, err
//        }
//        return state, nil
//    }
//    commonLen, _ := getCommonPrefix(seq, state.Sequence)
//    if commonLen < len(state.Sequence) {
//        return state, nil
//    }
//    children := state.getChildren()
//    nib := getNextNibble(commonLen, seq)
//    state.ChildKeys[nib] = getChildKey(key, nib)
//    _, err = del(state.ChildKeys[nib], seq[commonLen:])
//    if err != nil {
//        return nil, err
//    }
//    children[nib] = nil
//    if err == nil {
//        sum := 0
//        var uniqueChild *State
//        for _, child := range children {
//            if child != nil {
//                sum += 1
//                uniqueChild = child
//            }
//        }
//        if sum == 1 {
//            err = state.absorbOnlyChild(key, uniqueChild)
//            if err != nil {
//                return nil, err
//            }
//            return state, nil
//        }
//    }
//    state.resetValue(children)
//    err = db.putState(key, state)
//    if err != nil {
//        return nil, err
//    }
//    return state, nil
//}

//func getBackup(key []byte, seq string) (*State, error) {
//    state, err := db.getState(key)
//    if err != nil {
//        return nil, err
//    }
//    if state.IsLeaf {
//        if seq == state.Sequence {
//            return state, nil
//        }
//        return nil, errors.New("current key not found")
//    }
//    commonLen, _ := getCommonPrefix(seq, state.Sequence)
//    if commonLen < len(state.Sequence) {
//        return nil, errors.New("current key not found")
//    }
//    nib := getNextNibble(commonLen, seq)
//    return get(state.ChildKeys[nib], seq[commonLen:])
//}