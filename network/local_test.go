package network

import (
    "testing"
    "fmt"
    "net"
    "time"
)

func TestLocal(t *testing.T) {
    addr := &net.TCPAddr{Port: 55555}
    listener, err := net.ListenTCP("tcp", addr)
    if err != nil {
        fmt.Println("error", err)
        return
    }
    for {
        fmt.Println("start listen")
        conn, err := listener.AcceptTCP()
        fmt.Println("listen from ", conn.RemoteAddr())
        if err != nil {
            continue
        }
        b := make([]byte, 1024*1024)
        n, err := conn.Read(b)
        if err != nil {
            fmt.Println("err:", err)
            break
        } else {
            fmt.Println(n)
        }
        conn.Close()
    }
}

func TestLocal2(t *testing.T)  {
    //addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:55555")
    //fmt.Println(addr)
    //if err != nil {
    //    return
    //}
    b := make([]byte, 1024*10)
    for {
        //conn, err := net.DialTCP("tcp", nil, addr)

        d, err := time.ParseDuration("3s")
        if err != nil {
          fmt.Println(err)
          return
        }
        //add := "127.0.0.1:55555"
        add := "192.168.3.113:55556"
        conn, err := net.DialTimeout("tcp", add, d)

        //add := "192.168.3.113:55556"
        //conn, err := net.Dial("tcp", add) // This will block if it is offline, refused otherwise
        if err != nil {
            fmt.Printf("%T %v\n", err, err)
            if ope, ok := err.(*net.OpError); ok {
                fmt.Println(ope.Timeout(), ope)
            }
        }
        t := time.Now()
        d2, err := time.ParseDuration("1ms")
        if err != nil {
            fmt.Println(err)
        } else {
            conn.SetDeadline(t.Add(d2))
        }
        if err != nil {
            fmt.Println("error during dail:", err, "Fuck", err.Error())
            return
        } else {
            fmt.Println(conn)
        }
        if num, err := conn.Write(b); err != nil {
            fmt.Println("Send error ", err)
            return
        } else {
            fmt.Println("Send bytes ", num)
        }
        conn.Close()
    }
}