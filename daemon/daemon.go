package main

import (
    "log"
    "net"
    "net/http"
    "net/rpc"

    "github.com/mdlayher/wireguardctrl"
    "github.com/sevlyar/go-daemon"
)

const PORT = "7667"

func main() {
    // Setting up daemon
    ctx := &daemon.Context{
        PidFileName: "pid",
        PidFilePerm: 0644,
        LogFileName: "log",
        LogFilePerm: 0640,
        WorkDir:     "./", // TODO: change to /usr/local/bravod/ or /tmp/bravod/
        Umask:       027,
        Args:        []string{"bravod"},
    }

    d, err := ctx.Reborn()
    if err != nil {
        log.Fatal("Unable to run: ", err)
    }
    if d != nil {
        return
    }
    defer ctx.Release()

    // Setting up RPC
    done := make(chan interface{})
    go startRPC(done)
    <-done

    // TODO: graceful shutdown
}

func startRPC(done chan interface{}) {
    bravoRPC := new(BravoRPC)
    bravoRPC.w = NewWirectl()

    if err := rpc.Register(bravoRPC); err != nil {
        log.Fatalln("RPC Register error:", err)
    }
    rpc.HandleHTTP()

    log.Println("bravoRPC started")

    listener, err := net.Listen("tcp", ":" + PORT)
    if err != nil {
        log.Fatalln("Listen error:", err)
    }
    defer listener.Close()

    if err := http.Serve(listener, nil); err != nil {
        log.Fatalln("Serve error:", err)
    }

    done <- struct{}{} // perhaps no one will ever reach that point
}

func NewWirectl() *wireguardctrl.Client {
    c, err := wireguardctrl.New()
    if err != nil {
        log.Fatalln("Failed to open wireguardctrl: ", err)
    }

    return c
}
