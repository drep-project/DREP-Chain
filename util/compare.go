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

func SliceEqual(a []interface{}, b []interface{}, cp func(interface{}, interface{})bool) bool {
    if len(a) != len(b) {
        return false
    }
    l := len(a)
    for i := 0; i < l; i++ {
        if !cp(a[i], b[i]) {
            return false
        }
    }
    return true
}