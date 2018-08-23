package util

import "bytes"

func Contains(a []byte, b [][]byte) bool {
    length := len(b)
    for i := 0; i < length; i++ {
        if bytes.Equal(a, b[i]) {
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