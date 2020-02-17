package app

import (
	"os"
	"os/signal"
	"syscall"
)

func exitSignal(ch chan struct{}) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGSEGV, syscall.SIGTSTP)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGSEGV, syscall.SIGTSTP:
				ch <- struct{}{}
			default:
			}
		}
	}()
}
