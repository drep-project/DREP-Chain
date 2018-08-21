package test

import (
    "fmt"
    "time"
    "BlockChainTest/network"
)

func RemoteConnect(tag int) {
    switch tag {
    case 1:
        peer := network.GetPeer()
        peer.InitLeader()
        leader := peer.AsLeader
        leader.Listen()
        leader.Work()

        // step 1
        // leader request ticket
        err := leader.RequestTicket()
        fmt.Println("step 1:")
        fmt.Println("leader request ticket")
        fmt.Println("error: ", err)
        fmt.Println()

        var wait int64 = 0
        for wait < network.MaximumWaitTime.Nanoseconds() {
            wait += 1
            time.Sleep(time.Nanosecond)
        }

        if b := leader.ValidateTicket(); b {
            fmt.Println("leader validate ticket: ", b)
            fmt.Println()
        } else {
            return
        }
    case 0:
        peer := network.GetPeer()
        peer.InitMinor()
        minor := peer.AsMinor
        minor.Listen()
        minor.Work()
    }
}
