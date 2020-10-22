package d7024e

import(
	"encoding/hex"
    "encoding/json"
    "fmt"
    "net"
    "regexp"
    "strconv"
    "sync"
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
    store map[KademliaID][]byte
    mux sync.RWMutex
}

// Test channel
type TestPacket struct {
    b []byte
    n int
    addrStr string
}
var boTestMode bool
var testReadCh chan TestPacket
var testWriteCh chan TestPacket


func NewNetwork(ip string, port uint16) Network {
    id := NewRandomKademliaID()
    me := NewContact(id, "127.0.0.1:" + strconv.Itoa(int(port)))
    me.CalcDistance(me.ID)

    iface_port := ip + ":" + strconv.Itoa(int(port))
    boTestMode = false
    conn, err := net.ListenPacket("udp", iface_port)
    if err != nil {
        panic(err)
    }

    network := Network{
        RoutingTable: NewRoutingTable(me),
        MessageRecord: NewMessageRecord(),
        Conn: conn,
        store: make(map[KademliaID][]byte),
        mux: sync.RWMutex{},
    }

    go Listen(ip, port, &network)

    return network
}

func (network *Network) SetTestMode(readCh chan TestPacket, writeCh chan TestPacket) {
    boTestMode = true
    testReadCh = readCh
    testWriteCh = writeCh
}

func Listen(ip string, port uint16, network *Network) {
    var buf [4096]byte
    var addrStr string
    var n int
    var err error
    var addr net.Addr
    for {
        if boTestMode {
	    packet := <- testReadCh
            for i := 0 ; i < packet.n ; i++ {
                buf[i] = packet.b[i]
            }
            n = packet.n
            addrStr = packet.addrStr
            err = nil
        } else {
            n, addr, err = network.Conn.ReadFrom(buf[0:])
            addrStr = addr.String()
        }
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
////                Address: addr.String(),
                Address: addrStr,
            }
            network.RoutingTable.AddContact(contact)

            routeMsg(msg, &contact, network)
        } else {
            println("Received malformed message!")
        }
    }
}

func (network *Network) Join(peer string, ready chan bool) {
    requestID := NewRandomKademliaID()
    firstMsg := Message{
        Type : "RPC",
        Name : "PING",
        RequestID : requestID.String(),
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

    network.MessageRecord.RecordMessage(
        *requestID,
        func(msg Message, contact Contact) {
            fmt.Println(
                "RX_JOIN_RES:",
                msg.Source,
                msg.Destination,
                msg.RequestID,
            )
            network.SendFindContactMessage(
                *network.RoutingTable.me.ID,
                func() {},
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
    callback func(contact Contact),
) {
    requestID := NewRandomKademliaID()
    msg := Message{
        Type : "RPC",
        Name : "PING",
        RequestID : requestID.String(),
        Source : network.RoutingTable.me.ID.String(),
        Destination : contact.ID.String(),
        Params : []string{},
    }
    fmt.Println("TX_PING_RPC:", msg.Source, msg.Destination, msg.RequestID)
    network.send(msg, contact)

    network.MessageRecord.RecordMessage(
        *requestID,
        func(_ Message, contact Contact) {
            fmt.Println(
                "RX_PING_RES:",
                msg.Source,
                msg.Destination,
                msg.RequestID,
            )
            callback(contact)
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
    callback func(),
) {
    emptyMap := make(map[KademliaID]struct{})
    mux := sync.Mutex{}
    done := false
    network.sendFindContactMessage(askFor, &emptyMap, &mux, &done, callback)
}

func (network *Network) sendFindContactMessage(
    askFor KademliaID,
    sentTo *map[KademliaID]struct{},
    mux *sync.Mutex,
    done *bool,
    callback func(),
) {
    msg := Message{
        Type: "RPC",
        Name: "FIND-NODE",
        Source: network.RoutingTable.me.ID.String(),
        Params: []string{askFor.String()},
    }
    network.recurseMsg(
        &askFor,
        msg,
        sentTo,
        mux,
        done,
        "FIND",
        func(msg Message, contact Contact) {
            network.handleFindContactMessage(
                msg,
                &contact,
                askFor,
                sentTo,
                mux,
                done,
                callback,
            )
        },
        callback,
    )
}

func (network *Network) returnFindContactMessage(
    msg Message,
    contact *Contact,
) {
    fmt.Println("RX_FIND_RPC:", msg.Source, msg.Destination, msg.RequestID)
    if len(msg.Params) != 1 {
        println("Malformed FIND-NODE message!")
        return
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
    mux *sync.Mutex,
    done *bool,
    callback func(),
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

    network.sendFindContactMessage(askFor, sentTo, mux, done, callback)
}


func (network *Network) SendStoreMessage(data []byte, callback func()) {
    emptyMap := make(map[KademliaID]struct{})
    mux := sync.Mutex{}
    done := false
    network.sendStoreMessage(data, &emptyMap, &mux, &done, callback)
}

func (network *Network) sendStoreMessage(
    data []byte,
    sentTo *map[KademliaID]struct{},
    mux *sync.Mutex,
    done *bool,
    callback func(),
) {
    hash := Hash(data)
    msg := Message{
        Type: "RPC",
        Name: "STORE",
        Source: network.RoutingTable.me.ID.String(),
        Params: []string{hash.String(), hex.EncodeToString(data)},
    }
    network.recurseMsg(
        hash,
        msg,
        sentTo,
        mux,
        done,
        "STOR",
        func(msg Message, contact Contact) {
            network.handleStoreMessage(
                msg,
                &contact,
                data,
                sentTo,
                mux,
                done,
                callback,
            )
        },
        callback,
    )
}

func (network *Network) returnStoreMessage(msg Message, contact *Contact) {
    fmt.Println("RX_STOR_RPC:", msg.Source, msg.Destination, msg.RequestID)
    if len(msg.Params) != 2 {
        println("Malformed STORE message!")
        return
    }

    response := Message{
        Type: "RETURN",
        Name: "STORE",
        RequestID: msg.RequestID,
        Source: network.RoutingTable.me.ID.String(),
        Destination: contact.ID.String(),
        Params: []string{},
    }

    data, err := hex.DecodeString(msg.Params[1])
    if err != nil {
        println("%#v", err)
        return
    }
    hash := NewKademliaID(msg.Params[0])
    network.mux.Lock()
    network.store[*hash] = data
    network.mux.Unlock()

    fmt.Println("TX_STOR_RES:", msg.Source, msg.Destination, msg.RequestID)
    network.send(response, contact)
}

func (network *Network) handleStoreMessage(
    msg Message,
    contact *Contact,
    data []byte,
    sentTo *map[KademliaID]struct{},
    mux *sync.Mutex,
    done *bool,
    callback func(),
) {
    fmt.Println("RX_STOR_RES:", msg.Source, msg.Destination, msg.RequestID)
    params := msg.Params
    if len(params) != 0 {
        println("Malformed STORE response!")
        return
    }

    network.sendStoreMessage(data, sentTo, mux, done, callback)
}

func (network *Network) SendFindDataMessage(
    askFor KademliaID,
) {
    emptyMap := make(map[KademliaID]struct{})
    mux := sync.Mutex{}
    done := false
    network.sendFindDataMessage(askFor, &emptyMap, &mux, &done)
}

func (network *Network) sendFindDataMessage(
    askFor KademliaID,
    sentTo *map[KademliaID]struct{},
    mux *sync.Mutex,
    done *bool,
) {
    msg := Message{
        Type: "RPC",
        Name: "FIND-VALUE",
        Source: network.RoutingTable.me.ID.String(),
        Params: []string{askFor.String()},
    }
    network.recurseMsg(
        &askFor,
        msg,
        sentTo,
        mux,
        done,
        "FVAL",
        func(msg Message, contact Contact) {
            network.handleFindDataMessage(
                msg,
                &contact,
                askFor,
                sentTo,
                mux,
                done,
            )
        },
        func() {},
    )
}

func (network *Network) returnFindDataMessage(msg Message, contact *Contact) {
    fmt.Println("RX_FVAL_RPC:", msg.Source, msg.Destination, msg.RequestID)
    if len(msg.Params) != 1 {
        println("Malformed FIND-VALUE message!")
        return
    }

    response := Message{
        Type: "RETURN",
        Name: "FIND-VALUE",
        RequestID: msg.RequestID,
        Source: network.RoutingTable.me.ID.String(),
        Destination: contact.ID.String(),
    }

    asksFor := NewKademliaID(msg.Params[0])
    contacts := network.RoutingTable.FindClosestContacts(asksFor, bucketSize)

    ID := NewKademliaID(msg.Params[0])
    network.mux.RLock()
    value, ok := network.store[*ID]
    network.mux.RUnlock()
    if ok {
        response.Params = []string{
            hex.EncodeToString(value),
        }
    } else {
        contactList := make([]string, 0)
        for i := 0; i < len(contacts); i++ {
            contactList = append(contactList, contacts[i].ID.String())
            contactList = append(contactList, contacts[i].Address)
        }
        response.Params = contactList
    }

    fmt.Println("TX_FVAL_RES:", msg.Source, msg.Destination, msg.RequestID)
    network.send(response, contact)
}

func (network *Network) handleFindDataMessage(
    msg Message,
    contact *Contact,
    askFor KademliaID,
    sentTo *map[KademliaID]struct{},
    mux *sync.Mutex,
    done *bool,
) {
    fmt.Println("RX_FVAL_RES:", msg.Source, msg.Destination, msg.RequestID)
    params := msg.Params
    if len(params) == 0 || (len(params) % 2 == 1 && len(params) != 1) {
        println("Malformed FIND-VALUE response!")
        return
    }

    if len(params) == 1 {
        mux.Lock()
        if !*done {
            data, err := hex.DecodeString(msg.Params[0])
            if err != nil {
                println("%#v", err)
                return
            }
            println(string(data))
            *done = true
        }
        mux.Unlock()
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

    network.sendFindDataMessage(askFor, sentTo, mux, done)
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

    if boTestMode {
        var packet TestPacket
        packet.addrStr = contact.Address
        packet.b = serialized
        packet.n = len(serialized)
        testWriteCh <- packet
    } else {
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
}

func (network *Network) recurseMsg(
    askFor *KademliaID,
    origMsg Message,
    sentTo *map[KademliaID]struct{},
    mux *sync.Mutex,
    done *bool,
    log_name string,
    handler func(msg Message, contact Contact),
    callback func(),
) {
    sendTo := make([]*Contact, 0)
    searchLen := 1
    for searchLen <= bucketSize && len(sendTo) < ALPHA {
        closeSlice := network.RoutingTable.FindClosestContacts(
            askFor,
            searchLen,
        )

        if len(closeSlice) < searchLen {
            break
        }

        closeContact := closeSlice[len(closeSlice) - 1]
        mux.Lock()
        if _, ok := (*sentTo)[*closeContact.ID]; !ok {
            sendTo = append(sendTo, &closeContact)
            (*sentTo)[*closeContact.ID] = struct{}{}
        }
        mux.Unlock()
        searchLen++
    }

    // Base case
    if len(sendTo) == 0 {
        if !*done {
            *done = true
            callback()
        }
        return
    }

    for _, to := range sendTo {
        msg := origMsg
        requestID := NewRandomKademliaID()
        msg.RequestID = requestID.String()
        msg.Destination = to.ID.String()

        network.MessageRecord.RecordMessage(
            *requestID,
            handler,
            func() {
                network.recurseMsg(
                    askFor,
                    origMsg,
                    sentTo,
                    mux,
                    done,
                    log_name,
                    handler,
                    callback,
                )
            },
        )

        fmt.Println(
            "TX_" + log_name + "_RPC:",
            msg.Source,
            msg.Destination,
            msg.RequestID,
        )
        network.send(msg, to)
    }
}
