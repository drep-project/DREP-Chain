package list

type SortedLinkedList struct {
    head, tail *node
    len int
}

func NewSortedLinkedList() *LinkedList {
    return &LinkedList{len:0}
}

func (l *SortedLinkedList) Size() int {
    return l.len
}

func (l *SortedLinkedList) Push(e interface{}) {

}

func (l *SortedLinkedList) remove(n *node) {
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

func (l *SortedLinkedList) Remove(e interface{}, cp func(interface{}, interface{})bool) bool {
    for p := l.head; p != nil; p = p.next {
        if cp(p.value, e) {
            l.remove(p)
            return true
        }
    }
    return false
}

func (l *SortedLinkedList) iterator() *Iterator {
    return &Iterator{l:l}
}