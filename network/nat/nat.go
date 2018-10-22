package nat

import (
    "time"
    "net"
    "fmt"
)

var (
    // LAN IP ranges
    _, lan10, _  = net.ParseCIDR("10.0.0.0/8")
    _, lan176, _ = net.ParseCIDR("172.16.0.0/12")
    _, lan192, _ = net.ParseCIDR("192.168.0.0/16")
)

type Interface interface {
    AddMapping(protocol string, extport, intport int, name string, lifetime time.Duration) error
    DeleteMapping(protocol string, extport, intport int) error
    ExternalIP() (net.IP, error)
    String() string
}

func Map(protocol string, extport, intport int, name string) {
    ip, err := internalAddress()
    if err != nil {
        return
    }
    if lan10.Contains(ip) || lan176.Contains(ip) || lan192.Contains(ip) {
        m := Any()
        if err := m.AddMapping(protocol, extport, intport, name, 20 * time.Minute); err != nil {
            fmt.Println("Couldn't add port mapping", "err", err)
        } else {
            fmt.Println("Mapped network port")
        }
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

func internalAddress() (net.IP, error) {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return nil, err
    }
    for _, address := range addrs {
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                fmt.Println("ip:", ipnet.IP.String())
                return ipnet.IP, err
            }
        }
    }
    return nil, fmt.Errorf("get internal ip address error")
}
