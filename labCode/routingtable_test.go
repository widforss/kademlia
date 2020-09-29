package d7024e

import (
	"testing"
)

func cc(c1 []Contact, c2 []Contact) bool {
	if len(c1) != len(c2) {
		return false
	}
	for i := range c1 {
		if c1[i].String() != c2[i].String() {
			return false
		}
	}
	
	return true
}


// Test NewRoutingTable, AddContact, FindClosestContacts, Contains
func TestRoutingTable(t *testing.T) {
	id0 := NewKademliaID("0000000000000000000000000000000000000000")
	id1 := NewKademliaID("da39a3ee5e6b4b0d3255bfef95601890afd80709") // id<n> = sha1hex("")
	id2 := NewKademliaID("10a34637ad661d98ba3344717656fcc76209c2f8")
	id3 := NewKademliaID("3e6c06b1a28a035e21aa0a736ef80afadc43122c")
	//	id4 := NewKademliaID("3c7435cfd4e31b9be3991041c9a4f8292b752e5b")
	//	id5 := NewKademliaID("63027d7630360e4203c0e3f970ec2ffcfe5f8f1b")
	//	id6 := "ecc1978dca2e31d10751ede8d8753f1cbded832e")
	idf := NewKademliaID("ffffffffffffffffffffffffffffffffffffffff")
	//co0 := NewContact(id0, "localhost:8000")
	co1 := NewContact(id1, "localhost:8000")
	co2 := NewContact(id2, "localhost:8000")
	//	co3 := NewContact(id3, "localhost:8000")
	//	co4 := NewContact(id4, "localhost:8000")
	//	co5 := NewContact(id5, "localhost:8000")
	//	co6 := NewContact(id6, "localhost:8000")
	//	cof := NewContact(idf, "localhost:8000")

	// Table size 0 - { }  - look up non-existing - no contact returned
	rt := NewRoutingTable(co1)
	contacts := rt.FindClosestContacts(id0, 20)
	if !cc(contacts, []Contact{}) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}
	if rt.Contains(id1) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}

	// Table size 1 - {1} - look up the id - expect {1}
	rt.AddContact(co1)
	contacts = rt.FindClosestContacts(id1, 20)
	if !cc(contacts, []Contact{co1}) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}
	if !rt.Contains(id1) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}
	if rt.Contains(id2) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}

	// Table size 1 - {1} - look up id 0* - expect {1}
	contacts = rt.FindClosestContacts(id0, 20)
	if !cc(contacts, []Contact{co1}) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}
	if !rt.Contains(id1) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}

	// Table size 1 - {1} - look up id f* - expect {1}
	contacts = rt.FindClosestContacts(idf, 20)
	if !cc(contacts, []Contact{co1}) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}

	// Table size 2 - {12} - look up id 1 - expect {12}
	rt.AddContact(co2)
	contacts = rt.FindClosestContacts(id1, 20)
	if !cc(contacts, []Contact{co1,co2}) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}
	if !rt.Contains(id1) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}
	if !rt.Contains(id2) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}
	if rt.Contains(id3) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}

	// Table size 2 - {12} - look up id 2 - expect {21}
	contacts = rt.FindClosestContacts(id2, 20)
	if !cc(contacts, []Contact{co2,co1}) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}

	// Table size 2 - {12} - reaffirm 1 - look up id 1 - expect {12}
	rt.AddContact(co1)
	contacts = rt.FindClosestContacts(id1, 20)
	if !cc(contacts, []Contact{co1,co2}) {
		t.Errorf("TestMessage_ParseMessage: Error %d", len(contacts))
	}
}
