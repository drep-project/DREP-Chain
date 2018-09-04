package util

type ListNode struct {
    value  interface{}
    prev, next *ListNode
}

type LinkedList struct {
    head, tail *ListNode
    len int
}

func NewLinkedList() *LinkedList {
    return &LinkedList{len:0}
}

func (l *LinkedList) Size() int {
    return l.len
}

func (l *LinkedList) Push(e interface{}) {

}

func (l *LinkedList) remove(n *ListNode) {
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

type Iterator struct {
    l *LinkedList
    n *ListNode
}

func (i *Iterator) hasNext() bool {
    return i.n != nil
}

func (i *Iterator) next() interface{} {
    t := i.n
    i.n = i.n.next
    return t.value
}

func (i *Iterator) remove() {
    t := i.n
    i.l.remove(t)
    i.n = i.n.next
}