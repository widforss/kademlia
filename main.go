package main

import (
    kademlia "./labCode"
    "math/rand"
    "os"
    "time"
)

const PORT = 9000
const IFACE = "0.0.0.0"

func main() {
    rand.Seed(time.Now().UnixNano())
    if len(os.Args) != 1 + 1 {
        panic("Exactly one command line requirement (peer address) needed!")
    }
    kademlia.NewKademlia(IFACE, PORT, os.Args[1])
    <-make(chan struct{})
}

