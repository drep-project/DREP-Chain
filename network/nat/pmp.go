package nat

import (
    "net"
    "github.com/jackpal/go-nat-pmp"
    "github.com/jackpal/gateway"
    "time"
    "fmt"
    "strings"
)

type pmp struct {
    gw net.IP
    c  *natpmp.Client
}

func (n *pmp) AddMapping(protocol string, extport, intport int, name string, lifetime time.Duration) error {
    if lifetime <= 0 {
        return fmt.Errorf("lifetime must not be <= 0")
    }
    // Note order of port arguments is switched between our
    // AddMapping and the client's AddPortMapping.
    client := n.c
    _, err := client.AddPortMapping(strings.ToLower(protocol), intport, extport, int(lifetime/time.Second))
    return err
}

func (n *pmp) DeleteMapping(protocol string, extport, intport int) (err error) {
    // To destroy a mapping, send an add-port with an internalPort of
    // the internal port to destroy, an external port of zero and a
    // time of zero.
    _, err = n.c.AddPortMapping(strings.ToLower(protocol), intport, 0, 0)
    return err
}

func (n *pmp) ExternalIP() (net.IP, error) {
    response, err := n.c.GetExternalAddress()
    if err != nil {
        return nil, err
    }
    return response.ExternalIPAddress[:], nil
}

func (n *pmp) String() string {
    return fmt.Sprintf("NAT-PMP(%v)", n.gw)
}

func discoverPMP() *pmp {
    // run external address lookups on all potential gateways
    gatewayIP, err := gateway.DiscoverGateway()
    if err != nil {
        return nil
    }
    client := natpmp.NewClient(gatewayIP)

    if _, err := client.GetExternalAddress(); err != nil {
        return nil
    } else {
        fmt.Println("added pmp")
        return &pmp{gatewayIP, client}
    }
}

