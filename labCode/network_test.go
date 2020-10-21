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
	idf := NewKademliaID("ffffffffffffffffffffffffffffffffffffffff")
	//	co1 := NewContact(id1, "localhost:1111")
	co2 := NewContact(id2, "127.0.0.1:2222")

	var packet TestPacket
	
	nw := NewNetwork("localhost", 2000)
	readCh = make(chan TestPacket, 20)
	writeCh = make(chan TestPacket, 20)
	nw.SetTestMode(readCh, writeCh)


	//
	// ping test
	//

	//
	// 127.0.0.1:2000 (idNw) sends ping request to 127.0.0.1:2222 (id2)
	//

	nw.SendPingMessage(&co2, somecallback)
	println(len(writeCh))

	packet = <- writeCh
	msg, err := ParseMessage(packet.b)
	if err != nil {
		t.Errorf("unparsable msg");
	}

	idNwStr := msg.Source

	//
	// 127.0.0.1:2222 (id2) sends ping response to 127.0.0.1:2000 (idNw)
	//

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
	
	time.Sleep(TTL / 2)

	println(len(readCh))
	println(len(writeCh))
	if len(writeCh) != 0 || len(readCh) != 0 || !gotCallback {
		t.Errorf("ping error");
	}


	// (Contact co2 should be registered by nw)

	//
	// Test method Join
	//
	// 127.0.0.1:2000 (nwId) tries to join swarm by test 127.0.0.1:3333 (idf)
	//
	
	//
	// 127.0.0.1:2000 (nwId) sends ping to 127.0.0.1.3333 (idf)
	//

	readyCh := make(chan bool)

	nw.Join("127.0.0.1:3333", readyCh)

	time.Sleep(TTL / 2)
	
	packet = <- writeCh
	msg, err = ParseMessage(packet.b)
	if err != nil {
		t.Errorf("unparsable msg");
	}

	if msg.Type != "RPC" || msg.Name != "PING" || msg.Source != idNwStr || msg.Destination != idNwStr {
		t.Errorf("unparsable msg");
	}

	//
	// test 127.0.0.1:3333 (idf) sends ping response to 127.0.0.1:2000 (nwId)
	//

	reply = Message{
		Type : "RETURN",
		Name : "PING",
		RequestID : msg.RequestID,
		Source : idf.String(),
		Destination : msg.Source,
		Params : []string{},
	}
	serialized, err = json.Marshal(reply)

	packet.b = serialized
	packet.n = len(serialized)
	packet.addrStr = "127.0.0.1:3333"

	gotCallback = false
	readCh <- packet
	
	time.Sleep(TTL / 2)

	if len(readCh) != 0 {
		t.Errorf("ping error");
	}

	//
	// nw 127.0.0.1:2000 (nwId) sends find-node request to 127.0.0.1:3333 (idf)
	//                                                 and 127.0.0.1:2222 (id2)
	//

	if len(writeCh) != 2 {
		t.Errorf("find-contact error");
	}
	packet = <- writeCh
	msg, err = ParseMessage(packet.b)
	if err != nil {
		t.Errorf("unparsable msg");
	}

	if msg.Type != "RPC" || msg.Name != "FIND-NODE" || msg.Source != idNwStr || !(msg.Destination == idf.String() || msg.Destination == id2.String()) {
		t.Errorf("unparsable msg")
	}
	if msg.Type != "RPC" || msg.Name != "FIND-NODE" || msg.Source != idNwStr || !(msg.Destination == idf.String() || msg.Destination == id2.String()) {
		t.Errorf("unparsable msg")
	}
}

