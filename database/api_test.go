package database

import (
    "testing"
    "fmt"
)

func TestGetBlock(t *testing.T) {
    block := GetBlockOutsideTransaction(135)
    fmt.Println(block)
}
