package d7024e

import(
	"testing"
	"time"
	"encoding/json"
)

var callArray [2]struct {
	dest    string   // "succ" or "dead"
	message Message
	contact Contact
}
var calls int

// Test that msg1 equals msg2
func compareMsgs(msg1 Message, msg2 Message) bool {
	if msg1.Type != msg2.Type { return false }
	if msg1.Name != msg2.Name { return false }
	if msg1.RequestID != msg2.RequestID { return false }
	if msg1.Source != msg2.Source { return false }
	if msg1.Destination != msg2.Destination { return false }
	if len(msg1.Params) != len(msg2.Params) { return false }
	for i := range msg1.Params {
		if msg1.Params[i] != msg2.Params[i] { return false }
	}

	return true
}

func testAMessage(t *testing.T, msgIn Message) bool {
	serialized, errIn := json.Marshal(msgIn)
	msgOut, errOut := ParseMessage(serialized)
	if !compareMsgs(msgIn, msgOut) || errIn != nil || errOut != nil {
		return false
	}
	
	return true
}

func TestMessage_ParseMessage(t *testing.T) {
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("da39a3ee5e6b4b0d3255bfef95601890afd80709")
	id3 := NewKademliaID("ffffffffffffffffffffffffffffffffffffffff")
	co1 := NewContact(id1, "0.0.0.0:0")
	co2 := NewContact(id2, "localhost:1000")
	//testCo3 := NewContact(id3, "255.255.255.255:65535")
	
	params0 := []string { }
	params1 := []string {"55aa55aa55aa55aa55aa55aa55aa55aa55aa55aa"}	

	// type, name, requestID, 
	msgsSucc := []Message {
		{ "RPC",    "PING",       id3.String(), co1.String(), co2.String(), params0 },
		{ "RPC",    "FIND-NODE",  id3.String(), co1.String(), co2.String(), params1 },
		{ "RPC",    "FIND-VALUE", id3.String(), co1.String(), co2.String(), params1 },
		{ "RPC",    "STORE",      id3.String(), co1.String(), co2.String(), params1 },
		{ "RETURN", "PING",       id3.String(), co2.String(), co1.String(), params0 },
	}

	msgsFail := []Message {
		{ "",       "PING",       id3.String(), co1.String(), co2.String(), params0 },
		{ "RPC",    "",           id3.String(), co1.String(), co2.String(), params0 },
		{ "RPC",    "PING",       "x",          co1.String(), co2.String(), params0 },
		{ "RPC",    "PING",       id3.String(), "x",          co2.String(), params0 },
		{ "RPC",    "PING",       id3.String(), co1.String(), "x",          params0 },
	}

	// Break a message
	var serialized =[]byte("not json")
	_, errOut := ParseMessage(serialized)
	if errOut == nil {
		t.Errorf("TestMessages_ParseMessage: Error break a message")
	}

        // Test successful messages
	for i := range msgsSucc {
		if !testAMessage(t, msgsSucc[i]) {
			t.Errorf("TestMessage_ParseMessage: Error succ %d", i)
		}
	}
	
	for i := range msgsFail {
		if testAMessage(t, msgsFail[i]) {
			t.Errorf("TestMessages_ParseMessage: Error fail %d", i)
		}
	}
}

func testMessage_Success(message Message, contact Contact) {
	callArray[calls].dest = "succ"
	callArray[calls].message = message
	callArray[calls].contact = contact
	calls = calls + 1
}

func testMessage_Dead() {
	callArray[calls].dest = "dead"
	calls = calls + 1
}

func TestMessage_MessageProcess(t *testing.T) {
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("da39a3ee5e6b4b0d3255bfef95601890afd80709")  // sha1("")
	id3 := NewKademliaID("10a34637ad661d98ba3344717656fcc76209c2f8")  // sha1^2("")
	id4 := NewKademliaID("3e6c06b1a28a035e21aa0a736ef80afadc43122c")  // sha1^3("")
	co1 := NewContact(id1, "0.0.0.0:0")
	co2 := NewContact(id2, "localhost:1000")
	//testCo3 := NewContact(id3, "255.255.255.255:65535")
	
	params0 := []string { }
	//	params1 := []string {"55aa55aa55aa55aa55aa55aa55aa55aa55aa55aa"}	

	// type, name, requestID, 
	//	msgsSucc := []Message {
	//		{ "RPC",    "PING",       id3.String(), co1.String(), co2.String(), params0 },
	//		{ "RPC",    "FIND-NODE",  id3.String(), co1.String(), co2.String(), params1 },
	//		{ "RPC",    "FIND-VALUE", id3.String(), co1.String(), co2.String(), params1 },
	//		{ "RPC",    "STORE",      id3.String(), co1.String(), co2.String(), params1 },
	//		{ "RETURN", "PING",       id3.String(), co2.String(), co1.String(), params0 },
	//	}

	message := Message { "RPC",    "PING",       id3.String(), co1.String(), co2.String(), params0 }

	r := NewMessageRecord()

	// Test correct message
	calls = 0
	r.RecordMessage(*id3, testMessage_Success, testMessage_Dead)
	r.ActOnMessage(message, co1)
	time.Sleep(TTL * 2)
	if calls != 1 || callArray[0].dest != "succ" || callArray[0].contact != co1 || !compareMsgs(callArray[0].message, message) {
		t.Errorf("TestMessages_MessageProcess: Error 1")
	}

	// Test timeout
	calls = 0
	r.RecordMessage(*id3, testMessage_Success, testMessage_Dead)
	//	r.ActOnMessage(message, co1)
	time.Sleep(TTL * 2)
	if calls != 1 || callArray[0].dest != "dead" {
		t.Errorf("TestMessages_MessageProcess: Error 2")
	}

	// Test wrong message
	calls = 0
	r.RecordMessage(*id4, testMessage_Success, testMessage_Dead)
	r.ActOnMessage(message, co1)
	time.Sleep(TTL * 2)
	if calls != 1 || callArray[0].dest != "dead" {
		t.Errorf("TestMessages_MessageProcess: Error 3")
	}
}
