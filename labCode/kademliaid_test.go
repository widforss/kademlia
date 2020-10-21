package d7024e

import (
	"testing"
	"math/rand"
)

// Internal function to check that the string is lowercase hexadecimal
func checkHex(str string) bool {
	for j := 0; j < len(str); j++ {
		if !(str[j] >= '0' && str[j] <= '9' || str[j] >= 'a' && str[j] <= 'f') {
			return false;
		}
	}
	return true;
}

// Test that function NewKademliaID interprets the specified Kademlia ID correctly
// Test that function String outputs the same Kademlia ID as was put in
// Exclusions:
// * NewKademliaID assumes input of 20 characters. Only 20 characters tested.
func TestNewKademliaID(t *testing.T) {
	testValsStr := []string {
		"0000000000000000000000000000000000000000",
		"ffffffffffffffffffffffffffffffffffffffff",
		"123456789abcde23456789abcde3456789abc456",
		"23456789abcde3456789abcde456789abc56789a",
	}
	testValsObj := []KademliaID {
		{0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00},
		{0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff},
		{0x12,0x34,0x56,0x78,0x9a,0xbc,0xde,0x23,0x45,0x67,0x89,0xab,0xcd,0xe3,0x45,0x67,0x89,0xab,0xc4,0x56},
		{0x23,0x45,0x67,0x89,0xab,0xcd,0xe3,0x45,0x67,0x89,0xab,0xcd,0xe4,0x56,0x78,0x9a,0xbc,0x56,0x78,0x9a},
	}

	for i := 0; i < len(testValsStr); i++ {
		// 
		k := NewKademliaID(testValsStr[i])
		if *k != testValsObj[i] {
			t.Errorf("TestNewKademliaID: Error - Compare index %d", i)
		}

		str := testValsObj[i].String()
		if len(str) != 40 {
			t.Errorf("TestString: Error - Length index %d", i)
		}
		if str != testValsStr[i] {
			t.Errorf("TestString: Error - Compare index %d", i)
		}
		if !checkHex(str) {
			t.Errorf("TestString: Error - Hex index %d", i)
		}
	}
}

// Test that function Hash correctly converts sha512 of data to id correctly
// First 160 big-endian bits of the sha512 becomes the KademliaID
func TestCalcHash(t *testing.T) {
	testValsBytes := []([]byte) {
		{ },
		{ 48 },
	}
	testValsObj := []KademliaID {
		{0xcf,0x83,0xe1,0x35,0x7e,0xef,0xb8,0xbd,0xf1,0x54,0x28,0x50,0xd6,0x6d,0x80,0x07,0xd6,0x20,0xe4,0x05},  // Sha256("")
		{0x31,0xbc,0xa0,0x20,0x94,0xeb,0x78,0x12,0x6a,0x51,0x7b,0x20,0x6a,0x88,0xc7,0x3c,0xfa,0x9e,0xc6,0xf7},  // Sha256("0")
	}

	for i := 0; i < len(testValsBytes); i++ {
		// 
		k := Hash(testValsBytes[i])
		if *k != testValsObj[i] {
			t.Errorf("TestNewKademliaID: Error - Compare index %d", i)
		}

		str := testValsObj[i].String()
		if len(str) != 40 {
			t.Errorf("TestCalcHash: Error - Length index %d", i)
		}
		if !checkHex(str) {
			t.Errorf("TestCalcHash: Error - Hex index %d", i)
		}
	}
}

// Test that function Less compares correctly
func TestCalcLess(t *testing.T) {
	k01 := NewKademliaID("0000000000000000000000000000000000000001")
	k02 := NewKademliaID("0000000000000000000000000000000000000002")
	k40 := NewKademliaID("4000000000000000000000000000000000000000")
	k80 := NewKademliaID("8000000000000000000000000000000000000000")

	if k02.Less(k02) || k01.Less(k01) || k40.Less(k40) || k80.Less(k80) {
		t.Errorf("TestCalcDistance: Error - Selfless")
	}

	if !k01.Less(k02) || !k40.Less(k80) {
		t.Errorf("TestCalcDistance: Error - Compare")
	}
}

// Test that function Equals compares correctly
func TestCalcEquals(t *testing.T) {
	k01 := NewKademliaID("0000000000000000000000000000000000000001")
	k02 := NewKademliaID("0000000000000000000000000000000000000002")
	k40 := NewKademliaID("4000000000000000000000000000000000000000")
	k80 := NewKademliaID("8000000000000000000000000000000000000000")

	if !k02.Equals(k02) || !k01.Equals(k01) || !k40.Equals(k40) || !k80.Equals(k80) {
		t.Errorf("TestCalcDistance: Error - Identity")
	}

	if k01.Equals(k02) || k40.Equals(k80) {
		t.Errorf("TestCalcDistance: Error - Equality")
	}
}

// Test function CalcDistance calculation of XOR distance
func TestCalcDistance(t *testing.T) {
	k0 := NewKademliaID("0000000000000000000000000000000000000000")
	km := NewKademliaID("123456789abcde23456789abcde3456789abc456")
	kn := NewKademliaID("23456789abcde3456789abcde456789abc56789a")
	kx := NewKademliaID("317131f131713d6622ee226629b53dfd35fdbccc")  // Known xor(km, kn)
	
	dmm := km.CalcDistance(km)

	// Distance of km against itself is zero
	if !dmm.Equals(k0) {
		t.Errorf("TestCalcDistance: Error - Self distancing")
	}

	dnm := kn.CalcDistance(km)
	dmn := km.CalcDistance(kn)

	// Distance between km and kn is the same 
	if !dnm.Equals(dmn) {
		t.Errorf("TestCalcDistance: Error - Symmetry")
	}

	// Distance between n and m is the known XOR distance
	if !dnm.Equals(kx) {
		t.Errorf("TestCalcDistance: Error - Known distance")
	}
}

// Internal function to test function NewRandomKademliaID with a random seed
func testSeed(t *testing.T, seed int64) {
	rand.Seed(seed)
	k1 := NewRandomKademliaID()
	k2 := NewRandomKademliaID()
	if k1.String() == k2.String() {
		t.Errorf("TestNewRandomKademliaID: Error - Duplicated value seed %d", seed)
	}
	if !checkHex(k1.String()) || !checkHex(k2.String()) {
		t.Errorf("TestNewRandomKademliaID: Error - Hex value seed %d", seed)
	}
}

// Test function newRandomKademliaID
// Exclusions:
// * PRNG range/quality/length
func TestNewRandomKademliaID(t *testing.T) {
	testSeed(t, -(1 << 63))
	testSeed(t, -1)
	testSeed(t, 0)
	testSeed(t, 1)
	testSeed(t, (1 << 63) - 1)
}
