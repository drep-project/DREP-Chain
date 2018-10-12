package nat

import (
    "github.com/huin/goupnp"
    "time"
    "net"
    "fmt"
    "strings"
    "errors"
    "github.com/huin/goupnp/dcps/internetgateway1"
    "github.com/huin/goupnp/dcps/internetgateway2"
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
    return "UPNP " + n.service
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
    return nil, fmt.Errorf("Get internal ip address error!")
}

func discoverUPnP() Interface {
    found := make(chan *upnp, 2)
    // IGDv1
    go discover(found, internetgateway1.URN_WANConnectionDevice_1, func(dev *goupnp.RootDevice, sc goupnp.ServiceClient) *upnp {
        switch sc.Service.ServiceType {
        case internetgateway1.URN_WANIPConnection_1:
            return &upnp{dev, "IGDv1-IP1", &internetgateway1.WANIPConnection1{ServiceClient: sc}}
        case internetgateway1.URN_WANPPPConnection_1:
            return &upnp{dev, "IGDv1-PPP1", &internetgateway1.WANPPPConnection1{ServiceClient: sc}}
        }
        return nil
    })
    // IGDv2
    go discover(found, internetgateway2.URN_WANConnectionDevice_2, func(dev *goupnp.RootDevice, sc goupnp.ServiceClient) *upnp {
        switch sc.Service.ServiceType {
        case internetgateway2.URN_WANIPConnection_1:
            return &upnp{dev, "IGDv2-IP1", &internetgateway2.WANIPConnection1{ServiceClient: sc}}
        case internetgateway2.URN_WANIPConnection_2:
            return &upnp{dev, "IGDv2-IP2", &internetgateway2.WANIPConnection2{ServiceClient: sc}}
        case internetgateway2.URN_WANPPPConnection_1:
            return &upnp{dev, "IGDv2-PPP1", &internetgateway2.WANPPPConnection1{ServiceClient: sc}}
        }
        return nil
    })
    for i := 0; i < cap(found); i++ {
        if c := <-found; c != nil {
            return c
        }
    }
    return nil
}

// finds devices matching the given target and calls matcher for all
// advertised services of each device. The first non-nil service found
// is sent into out. If no service matched, nil is sent.
func discover(out chan<- *upnp, target string, matcher func(*goupnp.RootDevice, goupnp.ServiceClient) *upnp) {
    devs, err := goupnp.DiscoverDevices(target)
    if err != nil {
        out <- nil
        return
    }
    found := false
    for i := 0; i < len(devs) && !found; i++ {
        if devs[i].Root == nil {
            continue
        }
        devs[i].Root.Device.VisitServices(func(service *goupnp.Service) {
            if found {
                return
            }
            // check for a matching IGD service
            sc := goupnp.ServiceClient{
                SOAPClient: service.NewSOAPClient(),
                RootDevice: devs[i].Root,
                Location:   devs[i].Location,
                Service:    service,
            }
            sc.SOAPClient.HTTPClient.Timeout = RequestTimeout
            upnp := matcher(devs[i].Root, sc)
            if upnp == nil {
                return
            }
            // check whether port mapping is enabled
            if _, nat, err := upnp.client.GetNATRSIPStatus(); err != nil || !nat {
                return
            }
            out <- upnp
            found = true
        })
    }
    if !found {
        out <- nil
    }
}