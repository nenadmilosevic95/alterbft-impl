package consensus

import (
	"bytes"
	"fmt"
	"testing"

	"dslab.inf.usi.ch/tendermint/crypto"
)

func assertMessageEquals(m, expected *Message) error {
	if m == nil {
		return fmt.Errorf("Unexpected nil message")
	}
	if m.Type != expected.Type {
		return fmt.Errorf("Expected message type %v got %v\n", expected.Type, m.Type)
	}
	if m.Epoch != expected.Epoch {
		return fmt.Errorf("Expected message epoch %v got %v\n", expected.Epoch, m.Epoch)
	}
	if m.Height != expected.Height {
		return fmt.Errorf("Expected message height %v got %v\n", expected.Height, m.Height)
	}
	if !m.Block.Equal(expected.Block) {
		return fmt.Errorf("Expected message block %v got %v\n", expected.Block, m.Block)
	}
	if !m.BlockID.Equal(expected.BlockID) {
		return fmt.Errorf("Expected message blockID %v got %v\n", expected.BlockID, m.BlockID)
	}
	err := assertCertificateEquals(m.Certificate, expected.Certificate)
	if err != nil {
		return err
	}
	if m.Sender != expected.Sender {
		return fmt.Errorf("Expected message sender %v got %v\n", expected.Sender, m.Sender)
	}
	if !m.Signature.Equal(expected.Signature) {
		return fmt.Errorf("Expected message signature %v got %v\n", expected.Signature, m.Signature)
	}

	if m.Sender2 != expected.Sender2 {
		return fmt.Errorf("Expected message sender2 %v got %v\n", expected.Sender2, m.Sender2)
	}
	if !m.Signature2.Equal(expected.Signature2) {
		return fmt.Errorf("Expected message signature2 %v got %v\n", expected.Signature2, m.Signature2)
	}

	if (m.Type == DELTA_REQUEST || m.Type == DELTA_RESPONSE) && !bytes.Equal(m.payload, expected.payload) {
		return fmt.Errorf("Expected message payload %v got %v\n", expected.payload, m.payload)
	}
	return nil

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

func testMarshalling(m *Message) error {
	mMarshalled := m.Marshall()
	size := len(mMarshalled)
	if size != m.ByteSize() {
		return fmt.Errorf("Expected byte size %v got %v\n", m.ByteSize(), size)
	}
	mm := MessageFromBytes(mMarshalled)
	err := assertMessageEquals(mm, m)
	if err != nil {
		return err
	}
	return nil

}

func TestMessageMarshalling(t *testing.T) {
	keys := testGetKeys(3)
	// Test 1: Propose
	// A - with no certificate
	b0 := NewBlock(testRandValue(4), nil)
	m := NewProposeMessage(MIN_EPOCH, b0, nil, 0)
	m.Sign(keys[0])
	err := testMarshalling(m)
	if err != nil {
		t.Error(err)
	}
	// B - with certificate
	b1 := NewBlock(testRandValue(1024), b0)
	bc := testBlockCertificate(MIN_EPOCH, b0, 2)
	m = NewProposeMessage(MIN_EPOCH, b1, bc, 0)
	m.Sign(keys[0])
	err = testMarshalling(m)
	if err != nil {
		t.Error(err)
	}

	// Test 2: Silence
	m = NewSilenceMessage(MIN_EPOCH, 0)
	m.Sign(keys[0])
	err = testMarshalling(m)
	if err != nil {
		t.Error(err)
	}
	// Test 3: Vote
	m = NewVoteMessage(MIN_EPOCH, b0.BlockID(), b0.Height, 5, 0)
	b := make([]byte, SignatureSize)
	b[0] = '4'
	b[30] = '1'
	m.Signature2 = b
	m.Sign(keys[0])
	err = testMarshalling(m)
	if err != nil {
		t.Error(err)
	}
	// Test 4: Quit epoch
	//A - silence certificate
	sc := testSilenceCertificate(MIN_EPOCH)
	m = NewQuitEpochMessage(MIN_EPOCH, sc)
	m.Sign(keys[0])
	err = testMarshalling(m)
	if err != nil {
		t.Error(err)
	}
	//B - block certificate
	bc = testBlockCertificate(MIN_EPOCH, b0, 2)
	m = NewQuitEpochMessage(MIN_EPOCH, bc)
	m.Sign(keys[0])
	err = testMarshalling(m)
	if err != nil {
		t.Error(err)
	}

	m = NewCertificateMessage(MIN_EPOCH, bc)
	m.Sign(keys[0])
	err = testMarshalling(m)
	if err != nil {
		t.Error(err)
	}

	b = testRandValue(1024)
	b[0] = 0
	m = NewDeltaRequestMessage(b, 0)
	err = testMarshalling(m)
	if err != nil {
		t.Error(err)
	}
	m = NewDeltaResponseMessage(testRandValue(1024), 2)
	err = testMarshalling(m)
	if err != nil {
		t.Error(err)
	}

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
	err := assertMessageEquals(mm, m)
	if err != nil {
		t.Error(err)
	}
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
