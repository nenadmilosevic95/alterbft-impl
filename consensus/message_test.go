package consensus

import (
	"testing"

	"dslab.inf.usi.ch/tendermint/crypto"
)

func assertMessageEquals(t *testing.T, m, expected *Message) {
	t.Helper()
	if m == nil {
		t.Error("Unexpected nil message")
		return
	}
	if m.Type != expected.Type {
		t.Error("Expected message type \n", expected.Type, " got \n", m.Type)
	}
	if m.Epoch != expected.Epoch {
		t.Error("Expected message epoch \n", expected.Epoch, " got \n", m.Epoch)
	}
	if !m.Block.Equal(expected.Block) {
		t.Error("Expected message block \n", expected.Block, " got \n", m.Block)
	}
	if !m.BlockID.Equal(expected.BlockID) {
		t.Error("Expected message blockID \n", expected.BlockID, " got \n", m.BlockID)
	}
	assertCertificateEquals(t, m.Certificate, expected.Certificate)
	if m.Sender != expected.Sender {
		t.Error("Expected message sender \n", expected.Sender, " got \n", m.Sender)
	}
	if !m.Signature.Equal(expected.Signature) {
		t.Error("Expected message signature \n", expected.Signature, "got\n", m.Signature)
	}

}

func TestPayload(t *testing.T) {

}

func testGetKeys(n int) []crypto.PrivateKey {
	var keys []crypto.PrivateKey
	for i := 0; i < n; i++ {
		keys = append(keys, crypto.GeneratePrivateKey())
	}
	return keys
}

func testMarshalling(m *Message, t *testing.T) {
	mMarshalled := m.Marshall()
	size := len(mMarshalled)
	if size != m.ByteSize() {
		t.Error("Expected byte size", m.ByteSize(), "got", size)
	}
	mm := MessageFromBytes(mMarshalled)
	assertMessageEquals(t, mm, m)
}

func TestMessageMarshalling(t *testing.T) {
	keys := testGetKeys(3)
	// Test 1: Propose
	// A - with no certificate
	b0 := NewBlock(testRandValue(4), nil)
	m := NewProposeMessage(MIN_EPOCH, b0, nil, 0)
	m.Sign(keys[0])
	testMarshalling(m, t)
	// B - with certificate
	b1 := NewBlock(testRandValue(1024), b0)
	bc := testBlockCertificate(MIN_EPOCH, b0, 2)
	m = NewProposeMessage(MIN_EPOCH, b1, bc, 0)
	m.Sign(keys[0])
	testMarshalling(m, t)

	// Test 2: Silence
	m = NewSilenceMessage(MIN_EPOCH, 0)
	m.Sign(keys[0])
	testMarshalling(m, t)
	// Test 3: Vote
	m = NewVoteMessage(MIN_EPOCH, b0.BlockID(), 5)
	m.Sign(keys[0])
	testMarshalling(m, t)
	// Test 4: Quit epoch
	//A - silence certificate
	sc := testSilenceCertificate(MIN_EPOCH)
	m = NewQuitEpochMessage(MIN_EPOCH, sc)
	m.Sign(keys[0])
	testMarshalling(m, t)
	//B - block certificate
	bc = testBlockCertificate(MIN_EPOCH, b0, 2)
	m = NewQuitEpochMessage(MIN_EPOCH, bc)
	m.Sign(keys[0])
	testMarshalling(m, t)
}

func TestMessageSignatures(t *testing.T) {
	e := int64(3)
	var size int
	var m, mm *Message
	b := NewBlock(testRandValue(1024), nil)
	m = NewProposeMessage(e, b, nil, 0)
	priv := crypto.GeneratePrivateKey()
	m.Sign(priv)

	pub := priv.PubKey()
	if !m.VerifySignature(pub) {
		t.Errorf("Failed to verify signature \n%v\n%v\n", m.Signature, m.Payload())
	}
	pub2 := crypto.GeneratePrivateKey().PubKey()
	if m.VerifySignature(pub2) {
		t.Error("Signature verified with wrong key", m, pub2)
	}

	mBytes := m.Marshall()
	size = len(mBytes)
	if size != m.ByteSize() {
		t.Error("Expected byte size", m.ByteSize(), "got", len(mBytes))
	}
	mm = MessageFromBytes(mBytes)
	assertMessageEquals(t, mm, m)
	if !mm.VerifySignature(pub) {
		t.Errorf("Failed to verify signature \n%v\n%v\n%v\n%v\n", m.Signature, mm.Signature, m.payload, mm.payload)
	}
	if mm.VerifySignature(pub2) {
		t.Error("Signature verified with wrong key", mm, pub2)
	}
}

func BenchmarkMessageSigning(t *testing.B) {
	key := crypto.GeneratePrivateKey()
	b0 := NewBlock(testRandValue(128000), nil)
	b1 := NewBlock(testRandValue(128000), b0)
	bc := testBlockCertificate(MIN_EPOCH, b0, 2)
	m := NewProposeMessage(MIN_EPOCH, b1, bc, 0)
	t.ResetTimer()
	m.Sign(key)
}

func BenchmarkMessageVerify(t *testing.B) {
	key := crypto.GeneratePrivateKey()
	b0 := NewBlock(testRandValue(128000), nil)
	b1 := NewBlock(testRandValue(128000), b0)
	bc := testBlockCertificate(MIN_EPOCH, b0, 2)
	m := NewProposeMessage(MIN_EPOCH, b1, bc, 0)
	m.Sign(key)
	t.ResetTimer()
	m.VerifySignature(key.PubKey())
}
