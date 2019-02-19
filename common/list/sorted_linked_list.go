package list

type SortedLinkedList struct {
    LinkedList
    cp func(interface{}, interface{})int
}

func NewSortedLinkedList(cp func(interface{}, interface{})int) *SortedLinkedList {
    return &SortedLinkedList{LinkedList: *NewLinkedList(), cp:cp}
}

func (l *SortedLinkedList) Add(e interface{}) {
    l.len++
    n := &node{value: e}
    if l.head == nil {
        l.head = n
        l.tail = n
        return
    }
    var p *node
    for p = l.head; p != nil && l.cp(e, p.value) > 0; p = p.next {}
    if p == l.head {
        l.head.prev = n
        n.next = l.head
        l.head = n
    } else if p == nil {
        l.tail.next = n
        n.prev = l.tail
        l.tail = n
    } else {
        n.prev = p.prev
        n.next = p
        p.prev.next = n
        p.prev = n
    }
}