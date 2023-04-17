package consensus

import (
	"testing"
)

func TestSignatureEqual(t *testing.T) {
	b := testRandValue(SignatureSize)
	bb := testRandValue(SignatureSize)
	var s1, s2 Signature = nil, nil
	// Test 1: Both singatures are nil.
	if !s1.Equal(s2) {
		t.Errorf("Equal returns false for \n%v,\n%v", s1, s2)
	}
	// Test 2: First signature is nil.
	s2 = SignatureFromBytes(bb)
	if s1.Equal(s2) {
		t.Errorf("Equal returns true for \n%v,\n%v", s1, s2)
	}
	// Test 3: Second singature is nil.
	s2 = nil
	s1 = SignatureFromBytes(b)
	if s1.Equal(s2) {
		t.Errorf("Equal returns true for \n%v,\n%v", s1, s2)
	}
	// Test 4: Two same signatures.
	s1 = SignatureFromBytes(b)
	s2 = SignatureFromBytes(b)
	if !s1.Equal(s2) {
		t.Errorf("Equal returns false for \n%v,\n%v", s1, s2)
	}
	// Test 5: Two different signatures.
	s1 = SignatureFromBytes(b)
	s2 = SignatureFromBytes(bb)
	if s1.Equal(s2) {
		t.Errorf("Equal returns true for \n%v,\n%v", s1, s2)
	}
}

func TestSignatureMarshalling(t *testing.T) {
	// Test 1: Marshalling of signature.
	b := testRandValue(SignatureSize)
	s := SignatureFromBytes(b)
	buffer := make([]byte, s.ByteSize())
	s.MarshallTo(buffer)
	ss := SignatureFromBytes(buffer)
	if !s.Equal(ss) {
		t.Errorf("Marshalling of signature is not working expected \n%v, returned \n%v\n", s, ss)
	}
}
