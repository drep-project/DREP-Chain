package list

type node struct {
    value  interface{}
    prev, next *node
}

type removable interface {
    remove(node *node)
}

type Iterator struct {
    l removable
    n *node
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