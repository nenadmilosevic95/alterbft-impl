package consensus

import "testing"

func TestProposalSet(t *testing.T) {
	ps := NewProposalSet()
	b0 := NewBlock(testRandValue(1024), nil)
	bCert := NewBlockCertificate(MIN_EPOCH, b0.BlockID())
	b1a := NewBlock(testRandValue(1024), b0)
	b1b := NewBlock(testRandValue(1024), b0)
	b1c := NewBlock(testRandValue(1024), b0)
	ps.Add(NewProposeMessage(MIN_EPOCH+1, b1a, bCert, 0))
	if !ps.Has(b1a.BlockID()) || ps.Count() != 1 {
		t.Errorf("Expected true and 1 got false and %v", ps.Count())
	}
	ps.Add(NewProposeMessage(MIN_EPOCH+1, b1b, bCert, 0))
	if !ps.Has(b1b.BlockID()) || ps.Count() != 2 {
		t.Errorf("Expected true and 2 got false and %v", ps.Count())
	}
	ps.Add(NewProposeMessage(MIN_EPOCH+1, b1c, bCert, 0))
	if !ps.Has(b1c.BlockID()) || ps.Count() != 3 {
		t.Errorf("Expected true and 3 got false and %v", ps.Count())
	}
}

func TestCertificateSet(t *testing.T) {
	cs := NewCertificateSet()
	b0 := NewBlock(testRandValue(1024), nil)
	bc0 := cs.Get(MIN_EPOCH, b0.BlockID())
	if bc0 != nil {
		t.Errorf("Get should return nil not %v", bc0)
	}
	bc0 = NewBlockCertificate(MIN_EPOCH, b0.BlockID())
	cs.Add(bc0)
	if bc0 != cs.Get(bc0.Epoch, bc0.BlockID()) {
		t.Errorf("Get is not returning good certificate")
	}
}
