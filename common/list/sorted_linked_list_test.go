package list

import (
    "testing"
    "github.com/drep-project/drep-chain/common"
)

func TestSortedLinkedList(t *testing.T) {
    l := NewSortedLinkedList(func(a interface{}, b interface{}) int {
        a1, ok1 := a.(int)
        b1, ok2 := b.(int)
        if ok1 && ok2 {
            if a1 > b1 {
                return 1
            } else if a1 < b1 {
                return -1
            } else {
                return 0
            }
        } else {
            return 0
        }
    })
    l.Add(5)
    assert(t, common.SliceEqual(l.ToArray(), []interface{}{5}, cp))
    assert(t, common.SliceEqual(l.ToReverseArray(), []interface{}{5}, cp))
    assert(t, l.Size() == 1)
    l.Add(3)
    assert(t, common.SliceEqual(l.ToArray(), []interface{}{3, 5}, cp))
    assert(t, common.SliceEqual(l.ToReverseArray(), []interface{}{5, 3}, cp))
    assert(t, l.Size() == 2)
    l.Add(4)
    assert(t, common.SliceEqual(l.ToArray(), []interface{}{3, 4, 5}, cp))
    assert(t, common.SliceEqual(l.ToReverseArray(), []interface{}{5, 4, 3}, cp))
    assert(t, l.Size() == 3)
    l.Add(8)
    assert(t, common.SliceEqual(l.ToArray(), []interface{}{3, 4, 5, 8}, cp))
    assert(t, common.SliceEqual(l.ToReverseArray(), []interface{}{8, 5, 4, 3}, cp))
    assert(t, l.Size() == 4)
    l.Add(3)
    assert(t, common.SliceEqual(l.ToArray(), []interface{}{3, 3, 4, 5, 8}, cp))
    assert(t, common.SliceEqual(l.ToReverseArray(), []interface{}{8, 5, 4, 3, 3}, cp))
    assert(t, l.Size() == 5)
    l.Add(8)
    assert(t, common.SliceEqual(l.ToArray(), []interface{}{3, 3, 4, 5, 8, 8}, cp))
    assert(t, common.SliceEqual(l.ToReverseArray(), []interface{}{8, 8, 5, 4, 3, 3}, cp))
    assert(t, l.Size() == 6)
}