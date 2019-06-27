package messagepool

import (
	"fmt"
	"testing"
	"time"
)

func TestMessagePool_ObtainOne(t *testing.T) {
	p := New()
	p.Push(34)
	go func() {
		time.Sleep(1 * time.Second)
		p.Push(3)
		p.Push("oi")
		time.Sleep(1 * time.Second)
		p.Push(3)
	}()
	fmt.Println(p.ObtainOne(func(i interface{}) bool {
		if j, ok := i.(int); ok {
			return j == 3
		} else {
			return false
		}
	}, 3*time.Second))
	fmt.Println(p.ObtainOne(func(i interface{}) bool {
		if j, ok := i.(int); ok {
			return j == 3
		} else {
			return false
		}
	}, 3*time.Second))
	fmt.Println(p.ObtainOne(func(i interface{}) bool {
		_, ok := i.(string)
		return ok
	}, 3*time.Second))
	fmt.Println(p.ObtainOne(func(i interface{}) bool {
		if j, ok := i.(int); ok {
			return j == 4
		} else {
			return false
		}
	}, 3*time.Second))
}

func TestMessagePool_Obtain(t *testing.T) {
	p := New()
	p.Push(34)
	go func() {
		time.Sleep(1 * time.Second)
		p.Push(3)
		p.Push("oi")
		time.Sleep(1 * time.Second)
		p.Push(35)
		p.Push(36)
	}()
	fmt.Println(p.Obtain(2, func(i interface{}) bool {
		_, ok := i.(int)
		return ok
	}, 3*time.Second))
	fmt.Println(p.Obtain(1, func(i interface{}) bool {
		_, ok := i.(int)
		return ok
	}, 3*time.Second))
	fmt.Println(p.Obtain(2, func(i interface{}) bool {
		_, ok := i.(string)
		return ok
	}, 3*time.Second))
}

func TestMessagePool_Obtain2(t *testing.T) {
	p := New()
	p.Push(34)
	fmt.Println(p.Obtain(0, func(i interface{}) bool {
		_, ok := i.(int)
		return ok
	}, 3*time.Second))
}
