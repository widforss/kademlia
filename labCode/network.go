package d7024e

import(
    "encoding/json"
    "fmt"
    "net"
    "regexp"
    "strconv"
    "time"
)

const MESSAGE_RECORD_LEN = 1024
const ALPHA = 3
const TTL = time.Second

var hash_rx, _ = regexp.Compile(`[0-9a-f]{` + strconv.Itoa(IDLength) + `}`)

type Network struct {
    RoutingTable *RoutingTable
    MessageRecord MessageRecord
    Conn net.PacketConn
}

func NewNetwork(ip string, port uint16) Network {
    id := NewRandomKademliaID()
    me := NewContact(id, "127.0.0.1:" + strconv.Itoa(int(port)))
    me.CalcDistance(me.ID)

    iface_port := ip + ":" + strconv.Itoa(int(port))
    conn, err := net.ListenPacket("udp", iface_port)
    if err != nil {
        panic(err)
    }

    network := Network{
        RoutingTable: NewRoutingTable(me),
        MessageRecord: NewMessageRecord(),
        Conn: conn,
    }

    go Listen(ip, port, &network)

    return network
}

func Listen(ip string, port uint16, network *Network) {
    var buf [4096]byte
    for {
        n, addr, err := network.Conn.ReadFrom(buf[0:])
        if err != nil {
            println("%#v", err)
            continue
        }

        msg, err := ParseMessage(buf[:n])
        if err != nil {
            println("%#v", err)
            continue
        }

        meID := NewKademliaID(msg.Destination)
        sourceID := NewKademliaID(msg.Source)
        if meID.Equals(network.RoutingTable.me.ID) ||
            (msg.Type == "RPC" && msg.Name == "PING") {
            contact := Contact{
                ID: sourceID,
                Address: addr.String(),
            }
            network.RoutingTable.AddContact(contact)

            routeMsg(msg, &contact, network)
        } else {
            println("Received malformed message!")
        }
    }
}

func (network *Network) Join(peer string, ready chan bool) {
    msgID := NewRandomKademliaID().String()
    firstMsg := Message{
        Type : "RPC",
        Name : "PING",
        RequestID : msgID,
        Source : network.RoutingTable.me.ID.String(),
        Destination : network.RoutingTable.me.ID.String(),
        Params : []string{},
    }
    fmt.Println(
        "TX_JOIN_RPC:",
        firstMsg.Source,
        firstMsg.Destination,
        firstMsg.RequestID,
    )

    emptyMap := make(map[KademliaID]struct{})
    network.MessageRecord.RecordMessage(
        msgID,
        func(msg Message, contact Contact) {
            fmt.Println("RX_JOIN_RES:", msg.Source, msg.Destination, msg.RequestID)
            network.SendFindContactMessage(
                *network.RoutingTable.me.ID,
                &emptyMap,
            )
            ready <- true
        }, func() {
            ready <- false
        },
    )

    tmpContact := NewContact(NewRandomKademliaID(), peer)
    network.send(firstMsg, &tmpContact)
}

func (network *Network) SendPingMessage(
    contact *Contact,
    onReply func(contact Contact),
) {
    msgID := NewRandomKademliaID().String()
    msg := Message{
        Type : "RPC",
        Name : "PING",
        RequestID : msgID,
        Source : network.RoutingTable.me.ID.String(),
        Destination : contact.ID.String(),
        Params : []string{},
    }
    fmt.Println("TX_PING_RPC:", msg.Source, msg.Destination, msg.RequestID)
    network.send(msg, contact)

    network.MessageRecord.RecordMessage(
        msgID,
        func(_ Message, contact Contact) {
            fmt.Println("RX_PING_RES:", msg.Source, msg.Destination, msg.RequestID)
            onReply(contact)
        }, func() {
            println("PING RPC failed!")
        },
    )
}

func (network *Network) returnPingMessage(msg Message, contact *Contact) {
    fmt.Println("RX_PING_RPC:", msg.Source, msg.Destination, msg.RequestID)
    reply := Message{
        Type : "RETURN",
        Name : "PING",
        RequestID : msg.RequestID,
        Source : network.RoutingTable.me.ID.String(),
        Destination : contact.ID.String(),
        Params : []string{},
    }
    fmt.Println("TX_PING_RES:", msg.Source, msg.Destination, msg.RequestID)
    network.send(reply, contact)
}

func (network *Network) SendFindContactMessage(
    askFor KademliaID,
    sentTo *map[KademliaID]struct{},
) {
    sendTo := make([]*Contact, 0)
    searchLen := 1
    for searchLen >= bucketSize || len(sendTo) < ALPHA {
        closeSlice := network.RoutingTable.FindClosestContacts(
            &askFor,
            searchLen,
        )

        if len(closeSlice) < searchLen {
            break
        }

        closeContact := closeSlice[len(closeSlice) - 1]
        if _, ok := (*sentTo)[*closeContact.ID]; !ok {
            sendTo = append(sendTo, &closeContact)
            (*sentTo)[*closeContact.ID] = struct{}{}
        }
        searchLen++
    }

    // Base case
    if searchLen >= bucketSize || len(sendTo) == 0 {
        return
    }

    for _, to := range sendTo {
        msgID := NewRandomKademliaID().String()
        msg := Message{
            Type: "RPC",
            Name: "FIND-NODE",
            RequestID: msgID,
            Source: network.RoutingTable.me.ID.String(),
            Destination: to.ID.String(),
            Params: []string{askFor.String()},
        }

        network.MessageRecord.RecordMessage(
            msgID,
            func(msg Message, contact Contact) {
                network.handleFindContactMessage(msg, &contact, askFor, sentTo)
            }, func() {
                network.SendFindContactMessage(askFor, sentTo)
            },
        )

        fmt.Println("TX_FIND_RPC:", msg.Source, msg.Destination, msg.RequestID)
        network.send(msg, to)
    }
}

func (network *Network) returnFindContactMessage(
    msg Message,
    contact *Contact,
) {
    fmt.Println("RX_FIND_RPC:", msg.Source, msg.Destination, msg.RequestID)
    if len(msg.Params) != 1 {
        println("Malformed FIND-NODE message!")
    }

    response := Message{
        Type: "RETURN",
        Name: "FIND-NODE",
        RequestID: msg.RequestID,
        Source: network.RoutingTable.me.ID.String(),
        Destination: contact.ID.String(),
    }

    asksFor := NewKademliaID(msg.Params[0])
    contacts := network.RoutingTable.FindClosestContacts(asksFor, bucketSize)
    contactList := make([]string, 0)
    for i := 0; i < len(contacts); i++ {
        contactList = append(contactList, contacts[i].ID.String())
        contactList = append(contactList, contacts[i].Address)
    }
    response.Params = contactList
    fmt.Println("TX_FIND_RES:", msg.Source, msg.Destination, msg.RequestID)
    network.send(response, contact)
}

func (network *Network) handleFindContactMessage(
    msg Message,
    contact *Contact,
    askFor KademliaID,
    sentTo *map[KademliaID]struct{},
) {
    fmt.Println("RX_FIND_RES:", msg.Source, msg.Destination, msg.RequestID)
    params := msg.Params
    if len(params) == 0 || len(params) % 2 == 1 {
        println("Malformed FIND-NODE response!")
        return
    }

    var newContacts []Contact
    for i := 0; i < len(params); i = i + 2 {
        newID := params[i]
        newAddress := params[i + 1]
        newContact := NewContact(NewKademliaID(newID), newAddress)
        newContacts = append(newContacts, newContact)
        network.RoutingTable.AddContact(newContact)
    }

    network.SendFindContactMessage(askFor, sentTo)
}


func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}

func (network *Network) returnFindDataMessage(msg Message, contact *Contact) {
	// TODO
}

func (network *Network) returnStoreMessage(msg Message, contact *Contact) {
	// TODO
}

func (network *Network) handleFindDataMessage(msg Message, contact *Contact) {
	// TODO
}

func (network *Network) handleStoreMessage(msg Message, contact *Contact) {
    // TODO
}

func routeMsg(msg Message, contact *Contact, network *Network) {
    if msg.Type == "RPC" {
        switch msg.Name {
            case "PING":
                network.returnPingMessage(msg, contact)
            case "FIND-NODE":
                network.returnFindContactMessage(msg, contact)
            case "FIND-VALUE":
                network.returnFindDataMessage(msg, contact)
            case "STORE":
                network.returnStoreMessage(msg, contact)
        }
    } else if msg.Type == "RETURN" {
        network.MessageRecord.ActOnMessage(msg, *contact)
    }
}

func (network *Network) send(msg Message, contact *Contact) {
    serialized, err := json.Marshal(msg)
    if err != nil {
        println("%#v", err)
        return
    }

    addr, err := net.ResolveUDPAddr("udp", contact.Address)
    if err != nil {
        println("%#v", err)
        return
    }

    _, err = network.Conn.WriteTo(serialized, addr)
    if err != nil {
        println("%#v", err)
        return
    }
}

