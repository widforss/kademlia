package d7024e

import (
	"testing"
)

// Tests function NewContact and/by the way of function String
// Exclusions:
// * Lower functions assume input of 20 characters. Only 20 characters tested.
func TestNewContactAndString(t *testing.T) {
	testValsStrIn := []struct {
		id string
		addr string
	}{
		{ "0000000000000000000000000000000000000000", "0.0.0.0:0"             },
		{ "da39a3ee5e6b4b0d3255bfef95601890afd80709", "10.0.0.2:1111"         },  // sha1sum("", 0)
		{ "ffffffffffffffffffffffffffffffffffffffff", "255.255.255.255:99999" },
	}
	testValsStrOut := []string {
		"contact(\"0000000000000000000000000000000000000000\", \"0.0.0.0:0\")",
		"contact(\"da39a3ee5e6b4b0d3255bfef95601890afd80709\", \"10.0.0.2:1111\")",
		"contact(\"ffffffffffffffffffffffffffffffffffffffff\", \"255.255.255.255:99999\")",
	}

	for i := 0; i < len(testValsStrIn); i++ {
		k := NewKademliaID(testValsStrIn[i].id)
		c := NewContact(k, testValsStrIn[i].addr)
		if c.String() != testValsStrOut[i] {
			t.Errorf("TestNewContactAndString: Error index %d - compare", i)
		}
	}
}

// Tests functions CalcDistance, Less, Append, GetContacts, Swap, Len
// Exclusions:
// * GetContact has no range checking, it is assumed user calls with count=1..N
// * Swap has no range checking, it is assumed user calls with count=0..N-1
// * Sort, Less is assumed to do so by distance
func TestGeneral(t *testing.T) {
	k1  := NewKademliaID("0000000000000000000000000000000000000000")
	k2  := NewKademliaID("da39a3ee5e6b4b0d3255bfef95601890afd80709") // sha1sum("", 0)
	k3  := NewKademliaID("ffffffffffffffffffffffffffffffffffffffff")
	ck1 := NewContact(k1, "0.0.0.0:0")
	ck2 := NewContact(k2, "10.0.0.1:1000")
	ck3 := NewContact(k3, "255.255.255.255:65535")
	d12 := k1.CalcDistance(k2)  // da...
	d22 := k2.CalcDistance(k2)  // 00...
	d23 := k3.CalcDistance(k2)  // 25...

	// Test of CalcDistance
	ck1.CalcDistance(k2)
	ck2.CalcDistance(k2)
	ck3.CalcDistance(k2)
	if ck1.distance.String() != d12.String() ||
		ck2.distance.String() != d22.String() ||
		ck3.distance.String() != d23.String() {
		t.Errorf("TestContactGeneral: Error - CalcDistance")
	}

	// Test of Less
	if ck1.Less(&ck2) || !ck2.Less(&ck1) {
		t.Errorf("TestContactGeneral: Error - Less")
	}

	// Length of zero candidate list
	cc := ContactCandidates {}
	ca := []Contact {}
	if cc.Len() != 0 {
		t.Errorf("TestContactGeneral: Error - AddGetLen 1")
	}

	// Length of zero candidate list appended with contact vector
	cc.Append(ca)
	if cc.Len() != 0 {
		t.Errorf("TestContactGeneral: Error - AddGetLen 1")
	}

	// Length and content of zero candidate list appended with contact vector {ck3}
	ca = []Contact {ck3}
	cc.Append(ca)
	gc := cc.GetContacts(1)
	if cc.Len() != 1 || len(gc) != 1  || gc[0] != ck3 {
		t.Errorf("TestContactGeneral: Error - AddGetLen 1")
	}

	// Length and content of candidate list {ck3} appended with 2 element contact vector {ck2, ck1}
	ca = []Contact {ck2, ck1}
	cc.Append(ca)
	gc = cc.GetContacts(1)
	if cc.Len() != 3 || len(gc) != 1 ||
		gc[0] != ck3 {
		t.Errorf("TestContactGeneral: Error - AddGetLen 1")
	}
	gc = cc.GetContacts(3)
	if cc.Len() != 3 || len(gc) != 3 ||
		gc[0] != ck3 || gc[1] != ck2 || gc[2] != ck1 {
		t.Errorf("TestContactGeneral: Error - AddGetLen 2")
	}

	// Swap contact with itself and check results
	cc.Swap(0, 0)
	gc = cc.GetContacts(3)
	if cc.Len() != 3 || len(gc) != 3 ||
		gc[0] != ck3 || gc[1] != ck2 || gc[2] != ck1 {
		t.Errorf("TestContactGeneral: Error - AddGetLen 3")
	}

	// Swap two contacts 1 and 2, and check results
	cc.Swap(1, 2)
	gc = cc.GetContacts(3)
	if cc.Len() != 3 || len(gc) != 3 ||
		gc[0] != ck3 || gc[1] != ck1 || gc[2] != ck2 {
		t.Errorf("TestContactGeneral: Error - AddGetLen 4 ")
	}
	if (!cc.Less(2, 0) || cc.Less(0, 2)) {
		t.Errorf("TestContactGeneral: Error - AddGetLen 5")
	}
		
	cc.Sort()
	gc = cc.GetContacts(3)
	if cc.Len() != 3 || len(gc) != 3 ||
		gc[0] != ck2 || gc[1] != ck3 || gc[2] != ck1 {
		t.Errorf("TestContactGeneral: Error - Sort 1")
	}
}
