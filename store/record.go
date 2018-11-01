package store

var (
    forwardedTrans = make(map[string]bool)
    forwardedBlocks = make(map[string]bool)
)

func ForwardTransaction(id string) {
    forwardedTrans[id] = true
}

func ForwardedTransaction(id string) bool {
    v, exists := forwardedTrans[id]
    return v && exists
}

func ForwardBlock(id string) {
    forwardedBlocks[id] = true
}

func ForwardedBlock(id string) bool {
    v, exists := forwardedBlocks[id]
    return v && exists
}