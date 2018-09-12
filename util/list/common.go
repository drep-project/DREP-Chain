package list

type node struct {
    value  interface{}
    prev, next *node
}

type Iterator interface {
    HasNext() bool
    Next() interface{}
    Remove()
}

