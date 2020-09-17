package d7024e

import(
    "encoding/json"
    "fmt"
    "net"
    "regexp"
    "strconv"
)

var hash_rx, _ = regexp.Compile(`[0-9a-f]{` + strconv.Itoa(IDLength) + `}`)

type Network struct {
    RoutingTable *RoutingTable
}

type Message struct {
    Type string `json:"type"`
    Name string `json:"name"`
    RequestID string `json:"requestId"`
    Source string `json:"source"`
    Destination string `json:"destination"`
    Params []string `json:"params"`
}

func Listen(ip string, port int, network *Network) {
    iface_port := ip + ":" + strconv.Itoa(port)

    conn, err := net.ListenPacket("udp", iface_port)
    if err != nil {
            fmt.Println(err)
            return
    }
    defer conn.Close()

    var buf [4096]byte
    for {
        n, addr, err := conn.ReadFrom(buf[0:])
        if err != nil {
            fmt.Println(err)
            continue
        }

        msg, err := parseMsg(addr, buf[:n], network)
        if err != nil {
            fmt.Println(err)
            continue
        }

        id := NewKademliaID(msg.Source)
        contact := Contact{
            ID: id,
            Address: addr.String(),
            distance: network.RoutingTable.me.ID.CalcDistance(id),
        }
        network.RoutingTable.AddContact(contact)

        routeMsg(msg, &contact, network)
    }
}

// Below handlers for outgoing messages

func (network *Network) SendPingMessage(contact *Contact) {
    //testing required
    params := []string {}
    msg := Message{
        Type : "RPC",
        Name : "PING",
        RequestID : NewRandomKademliaID().String(),
        Source : network.RoutingTable.me.ID.String(),
        Destination : contact.ID.String(),
        Params : params,
    }
    err := sendMsg(msg, contact.Address)
    if err != nil {
        fmt.Println("Error when sending ping");
    }
    return
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}

// Below handlers for requests from other clients

func (network *Network) returnPingMessage(msg Message, contact *Contact) {
    //testing required
    params := []string {}
    reply := Message{
        Type : "RETURN",
        Name : "PING",
        RequestID : msg.RequestID,
        Source : network.RoutingTable.me.ID.String(),
        Destination : contact.ID.String(),
        Params : params,
    }
    err := sendMsg(reply, contact.Address)
    if err != nil {
        fmt.Println("Error when repling to ping");
    }
    return
}


func (network *Network) returnFindContactMessage(msg Message, contact *Contact) {
	// TODO
}

func (network *Network) returnFindDataMessage(msg Message, contact *Contact) {
	// TODO
}

func (network *Network) returnStoreMessage(msg Message, contact *Contact) {
	// TODO
}

// Below handlers for responses to messages we've sent

func (network *Network) handlePingMessage(msg Message, contact *Contact) {
	// TODO
}

func (network *Network) handleFindContactMessage(msg Message, contact *Contact) {
	// TODO
}

func (network *Network) handleFindDataMessage(msg Message, contact *Contact) {
	// TODO
}

func (network *Network) handleStoreMessage(msg Message, contact *Contact) {
    // TODO
}
/*
Converts msg to byte array then
sends byte array to target address at port 9000
returns error
*/
func sendMsg(msg Message, address string)(error) {
    //testing required
    buf, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("Serialization of msg failed")
    }
    var port = 9000
    iface_port := address + ":" +strconv.Itoa(port)
    conn, err := net.Dial("udp", iface_port)
    if err != nil {
        return fmt.Errorf("Failed to connect to host")
    }
    _, err = conn.Write(buf)
    if err != nil {
        return fmt.Errorf("Failed to write to host")
    }
    err = conn.Close()
    if err != nil {
        return fmt.Errorf("Error on close")
    }
    return nil
}

func parseMsg(addr net.Addr, buf []byte, network *Network) (Message, error) {
    var msg Message
    err := json.Unmarshal(buf, &msg)
    if err != nil {
        return msg, err
    }
    if msg.Type != "RPC" && msg.Type != "RETURN" {
        return msg, fmt.Errorf("Unknown message type: %q", msg.Type)
    }
    if  msg.Name != "PING"       && msg.Name != "FIND-NODE" &&
        msg.Name != "FIND-VALUE" && msg.Name != "STORE" {
        return msg, fmt.Errorf("Unknown message name: %q", msg.Name)
    }
    id_ok := hash_rx.MatchString(msg.RequestID)
    src_ok := hash_rx.MatchString(msg.Source)
    dst_ok := hash_rx.MatchString(msg.Destination)
    msg_ok := hash_rx.MatchString(msg.RequestID)
    if !id_ok || !src_ok || !dst_ok || !msg_ok {
        return msg, fmt.Errorf("Invalid hash in message:\n%+v", msg)
    }
    return msg, err
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
        switch msg.Name {
            case "PING":
                network.handlePingMessage(msg, contact)
            case "FIND-NODE":
                network.handleFindContactMessage(msg, contact)
            case "FIND-VALUE":
                network.handleFindDataMessage(msg, contact)
            case "STORE":
                network.handleStoreMessage(msg, contact)
        }
    }
}
