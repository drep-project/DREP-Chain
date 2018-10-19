package nat

import (
    "net"
    "time"
    "fmt"
)

type Interface interface {
    AddMapping(protocol string, extport, intport int, name string, lifetime time.Duration) error
    DeleteMapping(protocol string, extport, intport int) error
    ExternalIP() (net.IP, error)
    String() string
}

func Map(protocol string, extport, intport int, name string) {
    m := Any()
    if err := m.AddMapping(protocol, extport, intport, name, 20 * time.Minute); err != nil {
        fmt.Println("Couldn't add port mapping", "err", err)
    } else {
        fmt.Println("Mapped network port")
    }
}

func Any() Interface {
    return startautodisc("UPnP or NAT-PMP", func() Interface {
        found := make(chan Interface, 2)
        go func() { found <- discoverUPnP() }()
        go func() { found <- discoverPMP() }()
        for i := 0; i < cap(found); i++ {
            if c := <-found; c != nil {
                return c
            }
        }
        return nil
    })
}
