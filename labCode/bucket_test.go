package d7024e

import (
	"testing"
)

// Test functions NewContact, AddContact, Len, GetContactAndCalcDistance
// Exclusions:
// * AddContact does not tolerate nil input
// * GetContactAndCalcDistance does not tolerate nil input
func TestBucketGeneral(t *testing.T) {
	k1  := NewKademliaID("0000000000000000000000000000000000000000")
	k2  := NewKademliaID("da39a3ee5e6b4b0d3255bfef95601890afd80709") // sha1sum("", 0)
	k3  := NewKademliaID("ffffffffffffffffffffffffffffffffffffffff")
	ck1 := NewContact(k1, "0.0.0.0:0")
	ck2 := NewContact(k2, "10.0.0.1:1000")
	ck3 := NewContact(k3, "255.255.255.255:65535")
	d12 := k1.CalcDistance(k2)
	d22 := k2.CalcDistance(k2)
	d23 := k3.CalcDistance(k2)

	bucket := newBucket()

	// Test zero bucket
	if bucket.Len() != 0 {
		t.Errorf("TestBucketGeneral: Error 1");
	}

	// Test add contact 1 -> {1}
	bucket.AddContact(ck1)
	if bucket.Len() != 1 {
		t.Errorf("TestBucketGeneral: Error 2");
	}
	cc := bucket.GetContactAndCalcDistance(k2)
	if len(cc) != 1 || cc[0].String() != ck1.String() || *cc[0].distance != *d12 {
		t.Errorf("TestBucketGeneral: Error 3");
	}

	// Test re-add contact 1 -> {1}
	bucket.AddContact(ck1)
	if bucket.Len() != 1 {
		t.Errorf("TestBucketGeneral: Error 4");
	}
	cc = bucket.GetContactAndCalcDistance(k2)
	if len(cc) != 1 ||
		cc[0].String() != ck1.String() || *cc[0].distance != *d12 {
		t.Errorf("TestBucketGeneral: Error 5");
	}	
	
	// Test add contacts {2,3} -> {3,2,1}
	bucket.AddContact(ck2)
	bucket.AddContact(ck3)
	cc = bucket.GetContactAndCalcDistance(k2)
	if len(cc) != 3 ||
		cc[0].String() != ck3.String() || cc[1].String() != ck2.String() ||
		cc[2].String() != ck1.String() || *cc[0].distance != *d23 || *cc[1].distance != *d22 || *cc[2].distance != *d12 {
		t.Errorf("TestBucketGeneral: Error 6");
	}

	// Test re-add front contact -> no change {3,2,1}
	bucket.AddContact(ck3)
	cc = bucket.GetContactAndCalcDistance(k2)
	if len(cc) != 3 ||
		cc[0].String() != ck3.String() || cc[1].String() != ck2.String() || cc[2].String() != ck1.String() ||
		*cc[0].distance != *d23 || *cc[1].distance != *d22 || *cc[2].distance != *d12 {
		t.Errorf("TestGeneral: Error 7");
	}
	
	// Test re-add back contact (1) -> {1,3,2}
	bucket.AddContact(ck1)
	cc = bucket.GetContactAndCalcDistance(k2)
	if len(cc) != 3 ||
		cc[0].String() != ck1.String() || cc[1].String() != ck3.String() ||
		cc[2].String() != ck2.String() || *cc[0].distance != *d12 || *cc[1].distance != *d23 || *cc[2].distance != *d22 {
		t.Errorf("TestBucketGeneral: Error 8");
	}
}
