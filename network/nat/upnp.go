package nat

import (
    "fmt"
    "strings"
    "time"
    "net"
    "errors"
    "log"
    "github.com/huin/goupnp/dcps/internetgateway1"
    "github.com/huin/goupnp/dcps/internetgateway2"
    "github.com/huin/goupnp"
)

const RequestTimeout = 3 * time.Second

type upnp struct {
    dev     *goupnp.RootDevice
    service string
    client  upnpClient
}

type upnpClient interface {
    GetExternalIPAddress() (string, error)
    AddPortMapping(string, uint16, string, uint16, string, bool, string, uint32) error
    DeletePortMapping(string, uint16, string) error
    GetNATRSIPStatus() (sip bool, nat bool, err error)
}

func (n *upnp) AddMapping(protocol string, extport, intport int, desc string, lifetime time.Duration) error {
    ip, err := n.internalAddress()
    if err != nil {
        return nil
    }
    protocol = strings.ToUpper(protocol)
    lifetimeS := uint32(lifetime / time.Second)
    n.DeleteMapping(protocol, extport, intport)
    fmt.Printf("%s: adding port...\n", n)
    return n.client.AddPortMapping("", uint16(extport), protocol, uint16(intport), ip.String(), true, desc, lifetimeS)
}

func (n *upnp) DeleteMapping(protocol string, extport, intport int) error {
    return n.client.DeletePortMapping("", uint16(extport), strings.ToUpper(protocol))
}

func (n *upnp) ExternalIP() (addr net.IP, err error) {
    ipString, err := n.client.GetExternalIPAddress()
    if err != nil {
        return nil, err
    }
    ip := net.ParseIP(ipString)
    if ip == nil {
        return nil, errors.New("bad IP in response")
    }
    return ip, nil
}

func (n *upnp) String() string {
    return "UPnP " + n.service
}

func (n *upnp) internalAddress() (net.IP, error) {
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

func discoverUPnP() *upnp {
    found := make(chan *upnp, 5)

    go discover(found, internetgateway1.URN_WANConnectionDevice_1, internetgateway1.URN_WANIPConnection_1)
    go discover(found, internetgateway1.URN_WANConnectionDevice_1, internetgateway1.URN_WANPPPConnection_1)

    go discover(found, internetgateway2.URN_WANConnectionDevice_2, internetgateway2.URN_WANIPConnection_1)
    go discover(found, internetgateway2.URN_WANConnectionDevice_2, internetgateway2.URN_WANIPConnection_2)
    go discover(found, internetgateway2.URN_WANConnectionDevice_2, internetgateway2.URN_WANPPPConnection_1)

    fmt.Println("going to find somting... ")

    for i := 0; i < cap(found); i++ {
        if u := <-found; u != nil {
            fmt.Println("found upnp succeed")
            return u
        }
    }
    fmt.Println("found nothing!")
    return nil
}

func discover(found chan<- *upnp, devURNs string, srvURNs string) {
    switch devURNs {
    case internetgateway1.URN_WANConnectionDevice_1:
        switch srvURNs {
        case internetgateway1.URN_WANIPConnection_1:
            clients, errs, err := internetgateway1.NewWANIPConnection1Clients()
            if err != nil {
                return
            }
            for _, c := range clients {
                if c != nil {
                    dev := c.RootDevice
                    upnp := &upnp{dev,"IGDv1-IP1", c}
                    found <- upnp
                    display("IGDv1-IP1", errs, len(clients))
                    return
                }
            }
        case internetgateway1.URN_WANPPPConnection_1:
            clients, errs, err := internetgateway1.NewWANPPPConnection1Clients()
            if err != nil {
                return
            }
            for _, c := range clients {
                if c != nil {
                    dev := c.RootDevice
                    upnp := &upnp{dev, "IGDv1-PPP1", c}
                    found <- upnp
                    display("IGDv1-PPP1", errs, len(clients))
                    return
                }
            }
        }
    case internetgateway2.URN_WANConnectionDevice_2:
        switch srvURNs {
        case internetgateway2.URN_WANIPConnection_1:
            clients, errs, err := internetgateway2.NewWANIPConnection1Clients()
            if err != nil {
                return
            }
            for _, c := range clients {
                if c != nil {
                    dev := c.RootDevice
                    upnp := &upnp{dev, "IGDv2-IP1", c}
                    found <- upnp
                    display("IGDv2-IP1", errs, len(clients))
                    return
                }
            }
        case internetgateway2.URN_WANIPConnection_2:
            clients, errs, err := internetgateway2.NewWANIPConnection2Clients()
            if err != nil {
                return
            }
            for _, c := range clients {
                if c != nil {
                    dev := c.RootDevice
                    upnp := &upnp{dev, "IGDv2-IP2", c}
                    found <- upnp
                    display("IGDv2-IP2", errs, len(clients))
                    return
                }
            }
        case internetgateway2.URN_WANPPPConnection_1:
            clients, errs, err := internetgateway2.NewWANPPPConnection1Clients()
            if err != nil {
                return
            }
            for _, c := range clients {
                if c != nil {
                    dev := c.RootDevice
                    upnp := &upnp{dev, "IGDv2-PPP1", c}
                    found <- upnp
                    display("IGDv2-PPP1", errs, len(clients))
                    return
                }
            }
        }
    }
    found <- nil
}

func display(service string, errors []error, count int)  {
    log.Printf("%s: Got %d errors finding servers and %d successfully discovered.\n", service,
        len(errors), count)
    for i, e := range errors {
        log.Printf("%s: Error finding server #%d: %v\n",service, i+1, e)
    }
}



