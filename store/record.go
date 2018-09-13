package store

var (
    forwardedTrans = make(map[string]bool)
)

func Forward(id string) {
    forwardedTrans[id] = true
}

func Forwarded(id string) bool {
    v, exists := forwardedTrans[id]
    return v && exists
}
