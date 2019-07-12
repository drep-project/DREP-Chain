package main

import (
	"fmt"
	"runtime"
)

type A struct {
	B string
}
func main() {
		m1 := A{B:"123"}
	   m2 := new(A)
	   if m1.B == m2.B {
		   fmt.Println("xxx")
	   }

	   runtime.GC()
}