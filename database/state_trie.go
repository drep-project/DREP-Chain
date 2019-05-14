package database

import (
	"strconv"

	"github.com/drep-project/drep-chain/crypto/sha3"
)

type StateStore interface {
}

type State struct {
	Sequence  string
	ChildKeys [17][]byte
	Value     []byte
	IsLeaf    bool
	db        *Database `binary:"ignore"`
}

func (state *State) resetValue(children [17]*State) {
	stack := make([]byte, 32*17)
	for i := 0; i < 17; i++ {
		if children[i] != nil && children[i].Value != nil {
			copy(stack[i*17:(i+1)*17], children[i].Value)
		}
	}
	state.Value = sha3.Keccak256(stack)
}

func (state *State) getChildren() [17]*State {
	var children [17]*State
	for i := 0; i < 17; i++ {
		if state.ChildKeys[i] != nil {
			children[i], _ = state.db.GetState(state.ChildKeys[i])
		}
	}
	return children
}

func (state *State) getChildKey(key []byte, nib int) []byte {
	if state.ChildKeys[nib] != nil {
		return state.ChildKeys[nib]
	}
	return sha3.HashS256(key, []byte("child"), []byte(strconv.Itoa(nib)))
}

func newLeaf(db *Database, seq string, key, value []byte) (*State, error) {
	leaf := &State{
		Sequence: seq,
		Value:    value,
		IsLeaf:   true,
		db:       db,
	}
	err := db.PutState(key, leaf)
	if err != nil {
		return nil, err
	}
	return leaf, nil
}

func insert(db *Database, seq string, key, value []byte) (*State, error) {
	state, err := db.GetState(key)
	if err != nil {
		return newLeaf(db, seq, key, value)
	}
	prefix, offset := getCommonPrefix(seq, state.Sequence)
	if prefix == state.Sequence {
		if state.IsLeaf {
			if seq == state.Sequence {
				return state.insertExistedLeafValue(key, value)
			} else {
				return state.insertNewChildBranchOnLeaf(seq, offset, key, value)
			}
		} else {
			return state.insertProceedingOnCurrentBranch(seq, offset, key, value)
		}
	} else {
		return state.insertDivergingBranch(prefix, seq, offset, key, value)
	}
}

func (state *State) insertExistedLeafValue(key, value []byte) (*State, error) {
	state.Value = value
	err := state.db.PutState(key, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (state *State) insertNewChildBranchOnLeaf(seq string, offset int, key, value []byte) (*State, error) {
	var err error
	children := state.getChildren()

	state.ChildKeys[16] = state.getChildKey(key, 16)
	children[16], err = newLeaf(state.db, "", state.ChildKeys[16], state.Value)
	if err != nil {
		return nil, err
	}

	nib := getNextNibble(seq, offset)
	state.ChildKeys[nib] = state.getChildKey(key, nib)
	children[nib], err = newLeaf(state.db, seq[offset:], state.ChildKeys[nib], value)
	if err != nil {
		return nil, err
	}

	state.resetValue(children)
	state.IsLeaf = false
	err = state.db.PutState(key, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (state *State) insertProceedingOnCurrentBranch(seq string, offset int, key, value []byte) (*State, error) {
	var err error
	children := state.getChildren()

	nib := getNextNibble(seq, offset)
	state.ChildKeys[nib] = state.getChildKey(key, nib)
	children[nib], err = insert(state.db, seq[offset:], state.ChildKeys[nib], value)
	if err != nil {
		return nil, err
	}
	state.db.PutState(state.ChildKeys[nib], children[nib])
	if err != nil {
		return nil, err
	}

	state.resetValue(children)
	err = state.db.PutState(key, state)
	if err != nil {
		return nil, err
	}
	err = state.db.PutState(key, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (state *State) insertDivergingBranch(prefix, seq string, offset int, key, value []byte) (*State, error) {
	var err error
	children := state.getChildren()

	nib0 := getNextNibble(state.Sequence, offset)
	childKey0 := state.getChildKey(key, nib0)
	div := &State{}
	div.Sequence = state.Sequence[offset:]
	for i, child := range children {
		if state.ChildKeys[i] != nil {
			div.ChildKeys[i] = div.getChildKey(state.ChildKeys[nib0], i)
			err = state.db.PutState(div.ChildKeys[i], child)
			if err != nil {
				return nil, err
			}
			err = state.db.DelState(state.ChildKeys[i])
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
	err = state.db.PutState(childKey0, div)
	if err != nil {
		return nil, err
	}

	nib1 := getNextNibble(seq, offset)
	childKey1 := state.getChildKey(key, nib1)
	leaf, err := newLeaf(state.db, seq[offset:], childKey1, value)
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
	err = state.db.PutState(key, state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (state *State) del(key []byte, seq string) (*State, error) {
	state, err := state.db.GetState(key)
	if err != nil {
		return nil, err
	}
	if state.IsLeaf {
		if seq == state.Sequence {
			return state.delExistedLeaf(key)
		} else {
			return nil, ErrKeyNotFound
		}
	}
	return state.delProceedingOnCurrentBranch(seq, key)
}

func (state *State) delExistedLeaf(key []byte) (*State, error) {
	err := state.db.DelState(key)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (state *State) delProceedingOnCurrentBranch(seq string, key []byte) (*State, error) {
	_, offset := getCommonPrefix(seq, state.Sequence)
	if offset < len(state.Sequence) {
		return nil, ErrKeyNotFound
	}
	nib := getNextNibble(seq, offset)
	state.ChildKeys[nib] = state.getChildKey(key, nib)
	_, err := state.del(state.ChildKeys[nib], seq[offset:])
	if err != nil {
		return nil, err
	}
	childCount, onlyChild := state.countChildren()
	if childCount == 1 {
		return state.absorbOnlyChild(onlyChild, key)
	}
	return state, nil
}

func (state *State) countChildren() (int, *State) {
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

func (state *State) absorbOnlyChild(onlyChild *State, key []byte) (*State, error) {
	state.Sequence += onlyChild.Sequence
	state.Value = onlyChild.Value
	state.IsLeaf = onlyChild.IsLeaf
	state.ChildKeys = onlyChild.ChildKeys
	err := state.db.PutState(key, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (state *State) get(key []byte, seq string) (*State, error) {
	state, err := state.db.GetState(key)
	if err != nil {
		return nil, err
	}
	if state.IsLeaf {
		if seq == state.Sequence {
			return state, nil
		}
		return nil, ErrKeyNotFound
	}
	return state.getProceedingOnCurrentBranch(seq)
}

func (state *State) getProceedingOnCurrentBranch(seq string) (*State, error) {
	_, offset := getCommonPrefix(seq, state.Sequence)
	if offset < len(state.Sequence) {
		return nil, ErrKeyNotFound
	}
	nib := getNextNibble(seq, offset)
	return state.get(state.ChildKeys[nib], seq[offset:])
}

func (state *State) search(key []byte, seq string, depth int) {
	st, _ := state.db.GetState(key)
	seq += state.Sequence

	for i := 0; i < 17; i++ {
		if state.ChildKeys[i] != nil {
			st.search(st.ChildKeys[i], seq, depth+1)
		}
	}
}
