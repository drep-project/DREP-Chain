package concurrent

import (
    "time"
    "errors"
)

func ExecuteTimeoutTask(f func() interface{}, duration time.Duration) (interface{}, error) {
    ch := make(chan interface{}, 1)
    go func() {
        ch <- f()
    }()
    select {
    case m := <- ch:
        return m,nil
    case <-time.After(duration):
        return nil, errors.New("timeout")
    }
}
