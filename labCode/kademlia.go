package d7024e

import (
    "bufio"
    "os"
    "strings"
    "sync"
)

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
        go func() {
            reader := bufio.NewReader(os.Stdin)
            for {
                row, err := reader.ReadString('\n')
                if err != nil {
                    println("%#v", err)
                    continue
                }
                splitText := strings.Fields(row)

                if len(splitText) == 1 {
                    command := splitText[0]
                    switch command {
                        case "exit":
                            os.Exit(0)
                    }
                } else if len(splitText) > 1 {
                    command := splitText[0]
                    parameter := strings.Join(splitText[1:], " ")
                    switch command {
                        case "put":
                            kad.Store([]byte(parameter))
                        case "get":
                            kad.LookupData(parameter)
                    }
                }
            }
        } ()
        return kad
    }
    panic("Failed to join network!")
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
    kademlia.network.SendFindContactMessage(
        *target.ID,
        func() {},
    )
}

func (kademlia *Kademlia) LookupData(hash string) {
    hashID := NewKademliaID(hash)
    kademlia.network.SendFindDataMessage(*hashID)
}

func (kademlia *Kademlia) Store(data []byte) {
    hash := Hash(data)
    emptyMap := make(map[KademliaID]struct{})
    mux := sync.Mutex{}
    done := false
    kademlia.network.SendFindContactMessage(
        *hash,
        func() {
            kademlia.network.sendStoreMessage(
                data,
                &emptyMap,
                &mux,
                &done,
                func() {
                    println(hash.String())
                },
            )
        },
    )
}
