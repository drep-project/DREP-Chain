package objectemitter

import (
    "sync"
    "time"
)

type ObjectEmitter struct {
    objects []interface{}
    maxSize int
    duration time.Duration
    processor func([]interface{})
    finishChan chan struct{}
    fullChan chan struct{}
    lock sync.Mutex
}

func New(maxSize int, duration time.Duration, processor func([]interface{})) *ObjectEmitter {
    return &ObjectEmitter{
        objects:make([]interface{}, 0),
        maxSize:maxSize,
        duration:duration,
        processor:processor,
        finishChan:make(chan struct{}, 1),
        fullChan:make(chan struct{}, 1),
    }
}

func (e *ObjectEmitter) Push(o interface{}) {
    e.lock.Lock()
    defer e.lock.Unlock()
    e.objects = append(e.objects, o)
    e.trigger()
}

func (e *ObjectEmitter) trigger()  {
    if len(e.objects) >= e.maxSize {
        e.fullChan <- struct{}{}
    }
}

func (e *ObjectEmitter) Start()  {
    go func() {
        for {
            select {
            case <- time.After(e.duration): {
                e.lock.Lock()
                e.processor(e.objects)
                e.objects = make([]interface{}, 0)
                e.lock.Unlock()
            }
            case <- e.fullChan: {
                e.lock.Lock()
                e.processor(e.objects)
                e.objects = make([]interface{}, 0)
                e.lock.Unlock()
            }
            case <- e.finishChan:
                break
            }
        }
    }()
}

func (e *ObjectEmitter) Finish() {
    e.finishChan <- struct{}{}
}