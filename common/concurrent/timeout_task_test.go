package concurrent

import (
	"fmt"
	"testing"
	"time"
)

func TestExecuteTimeoutTask(t *testing.T) {
	fmt.Println(time.Now())
	fmt.Println(ExecuteTimeoutTask(func() interface{} {
		fmt.Println("OK1")
		time.Sleep(2 * time.Second)
		return nil
	}, 3*time.Second))
	fmt.Println(time.Now())
	fmt.Println(ExecuteTimeoutTask(func() interface{} {
		fmt.Println("OK2")
		time.Sleep(2 * time.Second)
		return nil
	}, 1*time.Second))
	fmt.Println(time.Now())
}

func TestNewCountDownLatch(t *testing.T) {
	l1 := NewCountDownLatch(3)
	fmt.Println(time.Now())
	//go func() {l.Done()}()
	//go func() {l.Done()}()
	//go func() {l.Done()}()
	l1.Cancel()
	l1.WaitTimeout(3 * time.Second)
	fmt.Println(time.Now())
	l2 := NewCountDownLatch(3)
	fmt.Println(time.Now())
	go func() { l2.Done() }()
	go func() { l2.Done() }()
	go func() { l2.Done() }()
	l2.WaitTimeout(3 * time.Second)
	l2.Done()
	fmt.Println(time.Now())
	l3 := NewCountDownLatch(3)
	fmt.Println(time.Now())
	go func() { l3.Done() }()
	go func() { l3.Done() }()
	l3.WaitTimeout(3 * time.Second)
	fmt.Println(time.Now())
}

func TestNewCountDownLatch2(t *testing.T) {
	l1 := NewCountDownLatch(3)
	fmt.Println(time.Now())
	//go func() {l.Done()}()
	//go func() {l.Done()}()
	//go func() {l.Done()}()
	l1.Cancel()
	l1.Wait()
	fmt.Println(time.Now())
	l2 := NewCountDownLatch(3)
	fmt.Println(time.Now())
	go func() { l2.Done() }()
	go func() { l2.Done() }()
	go func() { l2.Done() }()
	l2.Wait()
	fmt.Println(time.Now())
	l3 := NewCountDownLatch(3)
	fmt.Println(time.Now())
	go func() { l3.Done() }()
	go func() { l3.Done() }()
	l3.Wait()
	fmt.Println(time.Now())
}
