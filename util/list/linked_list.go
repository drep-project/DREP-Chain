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

func (l *LinkedList) iterator() *Iterator {
    return &Iterator{l:l}
}