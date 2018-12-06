package database

import (
    "testing"
    "fmt"
)

func TestGetBlock(t *testing.T) {
    block := GetBlock(135)
    fmt.Println(block)
}
