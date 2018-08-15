package util

func SliceEqual(a, b []byte) bool {
    if len(a) != len(b) {
        return false
    }
    length := len(a)
    for i := 0; i < length; i++ {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

func Contains(a []byte, b [][]byte) bool {
    length := len(b)
    for i := 0; i < length; i++ {
        if SliceEqual(a, b[i]) {
            return true
        }
    }
    return false
}

func Subset(a [][]byte, b [][]byte) bool {
    length := len(a)
    for i := 0; i < length; i++ {
        if !Contains(a[i], b) {
            return false
        }
    }
    return true
}