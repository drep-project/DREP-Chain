package nat

import (
    "time"
    "net"
    "fmt"
    "sync"
)

type autodisc struct {
    what string // type of interface
    once sync.Once
    doit func() Interface

    mu    sync.Mutex
    found Interface
}

func startautodisc(what string, doit func() Interface) Interface {
    return &autodisc{what: what, doit: doit}
}

func (n *autodisc) AddMapping(protocol string, extport, intport int, name string, lifetime time.Duration) error {
    if err := n.wait(); err != nil {
        return err
    }
    return n.found.AddMapping(protocol, extport, intport, name, lifetime)
}

func (n *autodisc) DeleteMapping(protocol string, extport, intport int) error {
    if err := n.wait(); err != nil {
        return err
    }
    return n.found.DeleteMapping(protocol, extport, intport)
}

func (n *autodisc) ExternalIP() (net.IP, error) {
    if err := n.wait(); err != nil {
        return nil, err
    }
    return n.found.ExternalIP()
}

func (n *autodisc) String() string {
    n.mu.Lock()
    defer n.mu.Unlock()
    if n.found == nil {
        return n.what
    } else {
        return n.found.String()
    }
}

func (n *autodisc) wait() error {
    n.once.Do(func() {
        n.mu.Lock()
        n.found = n.doit()
        n.mu.Unlock()
    })
    if n.found == nil {
        return fmt.Errorf("no %s router discovered", n.what)
    }
    return nil
}