package d7024e

import(
	"testing"
	"encoding/json"
	"time"
)

var readCh chan TestPacket
var writeCh chan TestPacket

var gotCallback bool

func testNwMsg_Success(message Message, contact Contact) {
	//	callArray[calls].dest = "succ"
	//	callArray[calls].message = message
	//	callArray[calls].contact = contact
	//	calls = calls + 1
}

func testNwMsg_Dead() {
	//	callArray[calls].dest = "dead"
	//	calls = calls + 1
}

func somecallback(contact Contact) {
	gotCallback = true
}

func TestNetwork(t *testing.T) {
	//	id0 := NewKademliaID("0000000000000000000000000000000000000000")
	//	id1 := NewKademliaID("1111111111111111111111111111111111111111")
	id2 := NewKademliaID("da39a3ee5e6b4b0d3255bfef95601890afd80709")
	//	idf := NewKademliaID("ffffffffffffffffffffffffffffffffffffffff")
	//	co1 := NewContact(id1, "localhost:1111")
	co2 := NewContact(id2, "127.0.0.1:2222")

	var packet TestPacket
	
	nw := NewNetwork("localhost", 2000)
	readCh = make(chan TestPacket, 20)
	writeCh = make(chan TestPacket, 20)
	nw.SetTestMode(readCh, writeCh)

	
	// ping test

	nw.SendPingMessage(&co2, somecallback)
	println(len(writeCh))

	packet = <- writeCh
	msg, err := ParseMessage(packet.b)
	if err != nil {
		t.Errorf("unparsable msg");
	}

	reply := Message{
		Type : "RETURN",
		Name : "PING",
		RequestID : msg.RequestID,
		Source : msg.Destination,
		Destination : msg.Source,
		Params : []string{},
	}
	serialized, err := json.Marshal(reply)

	packet.b = serialized
	packet.n = len(serialized)
	packet.addrStr = "127.0.0.1:2222"

	gotCallback = false
	readCh <- packet
	time.Sleep(TTL)

	println(len(readCh))
	println(len(writeCh))
	if len(writeCh) != 0 || len(readCh) != 0 || !gotCallback {
		t.Errorf("ping error");
	}

	
	// Test method Join

	//	readyCh := make(chan bool)

	//	nw.Join("127.0.0.1:3333", readyCh)

	//	packet = <- writeCh
	//	msg, err = ParseMessage(packet.b)
	//	if err != nil {
	//		t.Errorf("unparsable msg");
	//	}
}

