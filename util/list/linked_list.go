package list

type LinkedList struct {
    head, tail *node
    len int
}

func NewLinkedList() *LinkedList {
    return &LinkedList{len:0}
}

func (l *LinkedList) Size() int {
    return l.len
}

func (l *LinkedList) Push(e interface{}) {
    n := &node{value: e}
    if l.head == nil {
        l.head = n
        l.tail = n
    } else {
        l.tail.next = n
        n.prev = l.tail
        l.tail = n
    }
    l.len++
}

func (l *LinkedList) remove(n *node) {
    if l.head == n && l.tail == n {
        l.head = nil
        l.tail = nil
    } else if l.head == n {
        l.head = n.next
        n.next.prev = nil
        n.next = nil
    } else if l.tail == n {
        l.tail = n.prev
        n.prev.next = nil
        n.prev = nil
    } else {
        n.prev.next = n.next
        n.next.prev = n.prev
        n.next = nil
        n.prev = nil
    }
    l.len--
}

func (l *LinkedList) Remove(e interface{}, cp func(interface{}, interface{})bool) bool {
    for p := l.head; p != nil; p = p.next {
        if cp(p.value, e) {
            l.remove(p)
            return true
        }
    }
    return false
}

func (l *LinkedList) iterator() Iterator {
    return &linkedListIterator{l:l, n1:l.head}
}

func (l *LinkedList) ToArray() []interface{} {
    r := make([]interface{}, l.len)
    i := 0
    for p := l.head; p != nil; p = p.next {
        r[i] = p.value
        i++
    }
    return r
}

func (l *LinkedList) ToReverseArray() []interface{} {
    r := make([]interface{}, l.len)
    i := 0
    for p := l.tail; p != nil; p = p.prev {
        r[i] = p.value
        i++
    }
    return r
}

type linkedListIterator struct {
    l  *LinkedList
    n1, n2 *node // n2 should be nil
}

func (i *linkedListIterator) HasNext() bool {
    return i.n1 != nil
}

func (i *linkedListIterator) Next() interface{} {
    i.n2 = i.n1
    i.n1 = i.n1.next
    return i.n2.value
}

func (i *linkedListIterator) Remove() {
    i.l.remove(i.n2)
    i.n2 = nil
}