package log

import "fmt"

var DEBUG = false

func Println(a ...interface{})  {
    if DEBUG {
        fmt.Println(a)
    }
}

func Printf(format string, a ...interface{})  {
    if DEBUG {
        fmt.Printf(format, a)
    }
}

func Errorf(format string, a ...interface{})  {
    if DEBUG {
        fmt.Errorf(format, a)
    }
}