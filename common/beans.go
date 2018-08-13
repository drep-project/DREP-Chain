package common

const (
    MSG_BLOCK = 1
    MSG_TRANSACTION = 2
)
type Message struct {
    Type int
    Body interface{}
}

type Block struct {
    Id int
    Tran string
}

type Transaction struct {
    Id int
    Tran string
}
