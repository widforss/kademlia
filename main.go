package main

import (
    kademlia "./labCode"
    "strconv"
)

const PORT = 9000
const IFACE = "0.0.0.0"

func main() {
    id := kademlia.NewRandomKademliaID()
    me := kademlia.NewContact(id, "127.0.0.1:" + strconv.Itoa(PORT))
    network := kademlia.Network{
        RoutingTable: kademlia.NewRoutingTable(me),
    }
    kademlia.Listen(IFACE, PORT, &network)
}

