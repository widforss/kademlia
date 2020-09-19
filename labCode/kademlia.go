package d7024e

type Kademlia struct {
    network Network
}

func NewKademlia(ip string, port uint16, peer string) Kademlia {
    kad := Kademlia{
        network: NewNetwork(ip, port),
    }
    go Listen(ip, port, &kad.network)

    ready := make(chan bool)
    kad.network.Join(peer, ready)

    if <-ready {
        return kad
    }
    panic("Failed to join network!")
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
    emptyMap := make(map[KademliaID]struct{})
    kademlia.network.SendFindContactMessage(
        *target.ID,
        &emptyMap,
    )
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
