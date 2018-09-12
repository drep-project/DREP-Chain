package list

type ArrayList struct {
    array []interface{}
}

func NewArrayList() *ArrayList {
    return &ArrayList{array:make([]interface{}, 10)}
}

func (l *ArrayList) Size() int {
    return len(l.array)
}

func (l *ArrayList) Get(index int) interface{} {
    return l.array[index]
}

func (l *ArrayList) Add(e interface{}) {
    l.array = append(l.array, e)
}

func (l *ArrayList) Set(index int, e interface{}) {
    l.array[index] = e
}