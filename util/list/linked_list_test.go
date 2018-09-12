package list

import (
    "testing"
    "BlockChainTest/util"
)

var cp = func(a interface{}, b interface{}) bool {
    a1, ok1 := a.(int)
    b1, ok2 := b.(int)
    return ok1 && ok2 && a1 == b1;
}

func assert(t *testing.T, b bool)  {
    if !b {
        t.Error("Fail")
    }
}

func TestLinkedList_Common(t *testing.T) {
    l := NewLinkedList()
    for i := 0; i < 5; i++ {
        l.Add(i)
        if i + 1 != l.Size() {
            t.Error("fail", i)
        }
    }
    l.Remove(1, cp)
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{0, 2, 3, 4}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{4, 3, 2, 0}, cp))
    assert(t, l.Size() == 4)
    l.Remove(0, cp)
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{2, 3, 4}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{4, 3, 2}, cp))
    assert(t, l.Size() == 3)
    l.Remove(4, cp)
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{2, 3}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{3, 2}, cp))
    assert(t, l.Size() == 2)
    l.Remove(5, cp)
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{2, 3}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{3, 2}, cp))
    assert(t, l.Size() == 2)
    l.Add(5)
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{2, 3, 5}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{5, 3, 2}, cp))
    assert(t, l.Size() == 3)
    l.Remove(5, cp)
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{2, 3}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{3, 2}, cp))
    assert(t, l.Size() == 2)
    l.Remove(3, cp)
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{2}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{2}, cp))
    assert(t, l.Size() == 1)
    l.Remove(2, cp)
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{}, cp))
    assert(t, l.Size() == 0)
    l.Add(5)
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{5}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{5}, cp))
    assert(t, l.Size() == 1)
}

func TestLinkedList_Iterator1(t *testing.T) {
    l := NewLinkedList()
    for i:= 0; i < 5; i++ {
        l.Add(i)
    }
    it := func() []interface{} {
        var r []interface{}
        for i := l.Iterator(); i.HasNext(); {
            r = append(r, i.Next())
        }
        return r
    }
    remove := func(e int) {
        for i := l.Iterator(); i.HasNext(); {
            t := i.Next()
            if cp(t, e) {
                i.Remove()
            }
        }
    }
    assert(t, util.SliceEqual(it(), []interface{}{0, 1, 2, 3, 4}, cp))
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{0, 1, 2, 3, 4}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{4, 3, 2, 1, 0}, cp))
    remove(3)
    assert(t, util.SliceEqual(it(), []interface{}{0, 1, 2, 4}, cp))
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{0, 1, 2, 4}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{4, 2, 1, 0}, cp))
    remove(0)
    assert(t, util.SliceEqual(it(), []interface{}{1, 2, 4}, cp))
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{1, 2, 4}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{4, 2, 1}, cp))
    remove(4)
    assert(t, util.SliceEqual(it(), []interface{}{1, 2}, cp))
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{1, 2}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{2, 1}, cp))
    remove(1)
    assert(t, util.SliceEqual(it(), []interface{}{2}, cp))
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{2}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{2}, cp))
    remove(2)
    assert(t, util.SliceEqual(it(), []interface{}{}, cp))
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{}, cp))
}

func TestLinkedList_Iterator2(t *testing.T) {
    l := NewLinkedList()
    for i:= 0; i < 5; i++ {
        l.Add(i)
    }
    for i := l.Iterator(); i.HasNext(); {
        t := i.Next()
        if cp(t, 2) || cp(t, 3){
            i.Remove()
        }
    }
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{0, 1, 4}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{4, 1, 0}, cp))
    for i := l.Iterator(); i.HasNext(); {
        t := i.Next()
        if cp(t, 0) || cp(t, 4){
            i.Remove()
        }
    }
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{1}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{1}, cp))
}

func TestLinkedList_Iterator3(t *testing.T) {
    l := NewLinkedList()
    for i:= 0; i < 5; i++ {
        l.Add(i)
    }
    for i := l.Iterator(); i.HasNext(); {
        i.Next()
        i.Remove()
    }
    assert(t, util.SliceEqual(l.ToArray(), []interface{}{}, cp))
    assert(t, util.SliceEqual(l.ToReverseArray(), []interface{}{}, cp))
    assert(t, l.Size() == 0)
}