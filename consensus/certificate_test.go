package consensus

import (
	"bytes"
	"fmt"
	"testing"

	"dslab.inf.usi.ch/tendermint/crypto"
)

func assertCertificateEquals(t *testing.T, c, expected *Certificate) {
	t.Helper()
	if c == nil && expected == nil {
		return
	}
	if c == nil && expected != nil || c != nil && expected == nil {
		t.Error("Unexpected nil message")
		return
	}
	if c.Type != expected.Type {
		t.Error("Expected certificate type \n", expected.Type, " got \n", c.Type)
	}
	if c.Epoch != expected.Epoch {
		t.Error("Expected certificate epoch \n", expected.Epoch, " got \n", c.Epoch)
	}
	if c.Height != expected.Height {
		t.Error("Expected certificate height \n", expected.Height, " got \n", c.Height)
	}
	if !c.BlockID().Equal(expected.BlockID()) {
		t.Error("Expected certificate blockID \n", expected.BlockID(), " got \n", c.BlockID())
	}
	for i := 0; i < len(c.Signatures); i++ {
		if !c.Signatures[i].Equal(expected.Signatures[i]) {
			t.Errorf("Expected signature %v for sender %v\n, got %v.\n", expected.Signatures[i], i, c.Signatures[i])
		}
	}
}

func testBlockCertificate(e int64, block *Block, n int) *Certificate {
	bc := NewBlockCertificate(e, block.BlockID(), block.Height)
	keys := testGetKeys(n)
	votes := make([]*Message, n)
	for i := 0; i < len(votes); i++ {
		votes[i] = NewVoteMessage(e, block.BlockID(), block.Height, int16(i))
		votes[i].Sign(keys[i])
		bc.AddSignature(votes[i].Signature, votes[i].Sender)
	}
	return bc
}

func testSilenceCertificate(e int64) *Certificate {
	sc := NewSilenceCertificate(e)
	keys := testGetKeys(2)
	silences := make([]*Message, 2)
	for i := 0; i < len(silences); i++ {
		silences[i] = NewSilenceMessage(e, int16(i))
		silences[i].Sign(keys[i])
		sc.AddSignature(silences[i].Signature, silences[i].Sender)
	}
	return sc
}

func TestEmptyBlockCertificate(t *testing.T) {
	e := int64(7)
	block := NewBlock(testRandValue(1024), nil)
	c := NewBlockCertificate(e, block.BlockID(), block.Height)
	if c == nil {
		t.Fatal("Nil certificate")
	}
	if c.Type != BLOCK_CERT {
		t.Error("Wrong certificate Type", c.Type, BLOCK_CERT)
	}
	if c.Epoch != e {
		t.Error("Wrong certificate Epoch", c.Epoch, e)
	}
	if !c.blockID.Equal(block.BlockID()) {
		t.Error("Wrong certificate blockID", c.blockID, block.BlockID())
	}
	if len(c.Signatures) != 0 {
		t.Error("Expected no Signatures, got", c.Signatures)
	}
	messages := c.ReconstructMessages()
	if len(messages) != 0 {
		t.Error("Expected no messages, got", messages)
	}

	b := c.Marshall()
	if len(b) != c.ByteSize() {
		t.Error("Marshalled size differs ByteSize:", len(b), c.ByteSize())
	}

	c = CertificateFromBytes(b)
	if c == nil {
		t.Fatal("Nil unmarshalled certificate")
	}
	if c.Type != BLOCK_CERT {
		t.Error("Wrong unmarshalled certificate Type", c.Type, BLOCK_CERT)
	}
	if c.Epoch != e {
		t.Error("Wrong unmarshalled certificate Epoch", c.Epoch, e)
	}
	if !c.blockID.Equal(block.BlockID()) {
		t.Error("Wrong unmarshalled certificate blockID", c.blockID, block.BlockID())
	}
	if len(c.Signatures) != 0 {
		t.Error("Expected no Signatures, got", c.Signatures)
	}
	messages = c.ReconstructMessages()
	if len(messages) != 0 {
		t.Error("Expected no messages, got", messages)
	}
}

func TestEmptySilenceCertificate(t *testing.T) {
	e := int64(7)
	c := NewSilenceCertificate(e)
	if c == nil {
		t.Fatal("Nil certificate")
	}
	if c.Type != SILENCE_CERT {
		t.Error("Wrong certificate Type", c.Type, SILENCE_CERT)
	}
	if c.Epoch != e {
		t.Error("Wrong certificate Epoch", c.Epoch, e)
	}
	// FIXME: which height of a silence certificate???
	//	if c.height != h {
	//		t.Error("Wrong certificate Height", c.height, h)
	//	}
	if c.blockID != nil {
		t.Error("Unexpected certificate blockID", c.blockID)
	}
	if len(c.Signatures) != 0 {
		t.Error("Expected no Signatures, got", c.Signatures)
	}
	messages := c.ReconstructMessages()
	if len(messages) != 0 {
		t.Error("Expected no messages, got", messages)
	}

	b := c.Marshall()
	if len(b) != c.ByteSize() {
		t.Error("Marshalled size differs ByteSize:", len(b), c.ByteSize())
	}

	c = CertificateFromBytes(b)
	if c == nil {
		t.Fatal("Nil unmarshalled certificate")
	}
	if c.Type != SILENCE_CERT {
		t.Error("Wrong unmarshalled certificate Type", c.Type, SILENCE_CERT)
	}
	if c.Epoch != e {
		t.Error("Wrong unmarshalled certificate Epoch", c.Epoch, e)
	}
	// FIXME: which height of a silence certificate???
	//	if c.height != h {
	//		t.Error("Wrong unmarshalled certificate height", c.height, h)
	//	}
	if c.blockID != nil {
		t.Error("Unexpected certificate blockID", c.blockID)
	}
	if len(c.Signatures) != 0 {
		t.Error("Expected no Signatures, got", c.Signatures)
	}
	messages = c.ReconstructMessages()
	if len(messages) != 0 {
		t.Error("Expected no messages, got", messages)
	}
}

func findMessageCertificate(t *testing.T, m *Message, c *Certificate, count int) {
	t.Helper()
	// Proper number of Senders and Signatures
	if len(c.Signatures) != count {
		t.Fatal("Expected", count, " Signatures, got",
			len(c.Signatures))
	}

	if c.Signatures[m.Sender] == nil {
		t.Fatal("Couldn't find sender", m.Sender, c.Signatures)
	}
	if !c.Signatures[m.Sender].Equal(m.Signature) {
		t.Error("Wrong signature for sender", m.Sender,
			m.Signature, c.Signatures[m.Sender])
	}
	index := -1
	messages := c.ReconstructMessages()
	//	payloads := c.ReconstructAndMarshallMessages()
	for i := range messages {
		if messages[i].Sender == m.Sender {
			index = i
			break
		}
	}
	if index < 0 {
		t.Error("Couldn't find reconstructed message with sender",
			m.Sender, messages)
	} else {
		assertMessageEquals(t, messages[index], m)
		//		m := NewMessageFromBytes(payloads[index])
		//		assertMessageEquals(t, m, v)
	}
}

func TestCertificateWithVotes(t *testing.T) {
	e := int64(7)
	// FIXME: where the height comes from?
	block := NewBlock(testRandValue(1024), nil)
	c := NewBlockCertificate(e, block.BlockID(), block.Height)

	v1 := NewVoteMessage(e, block.BlockID(), block.Height, 0)
	v1.Sign(crypto.GeneratePrivateKey())
	c.AddSignature(v1.Signature, v1.Sender)
	findMessageCertificate(t, v1, c, 1)

	b := c.Marshall()
	if len(b) != c.ByteSize() {
		t.Error("Marshalled size differs ByteSize:", len(b), c.ByteSize())
	}
	c = CertificateFromBytes(b)
	findMessageCertificate(t, v1, c, 1)

	v2 := NewVoteMessage(e, block.BlockID(), block.Height, 3)
	v2.Sign(crypto.GeneratePrivateKey())
	c.AddSignature(v2.Signature, v2.Sender)
	fmt.Print(c)
	findMessageCertificate(t, v1, c, 2)
	findMessageCertificate(t, v2, c, 2)

	b = make([]byte, c.ByteSize())
	c.MarshallTo(b)
	c = CertificateFromBytes(b)
	findMessageCertificate(t, v1, c, 2)
	findMessageCertificate(t, v2, c, 2)

	v3 := NewVoteMessage(e, block.BlockID(), block.Height, 27)
	v3.Sign(crypto.GeneratePrivateKey())
	c.AddSignature(v3.Signature, v3.Sender)
	findMessageCertificate(t, v1, c, 3)
	findMessageCertificate(t, v2, c, 3)
	findMessageCertificate(t, v3, c, 3)

	b = make([]byte, c.ByteSize())
	c.MarshallTo(b)
	c = CertificateFromBytes(b)
	findMessageCertificate(t, v1, c, 3)
	findMessageCertificate(t, v2, c, 3)
	findMessageCertificate(t, v3, c, 3)
}

func TestCertificateWithSilences(t *testing.T) {
	e := int64(7)
	c := NewSilenceCertificate(e)

	v1 := NewSilenceMessage(e, 0)
	v1.Sign(crypto.GeneratePrivateKey())
	c.AddSignature(v1.Signature, v1.Sender)
	findMessageCertificate(t, v1, c, 1)

	b := c.Marshall()
	if len(b) != c.ByteSize() {
		t.Error("Marshalled size differs ByteSize:", len(b), c.ByteSize())
	}
	c = CertificateFromBytes(b)
	findMessageCertificate(t, v1, c, 1)

	v2 := NewSilenceMessage(e, 3)
	v2.Sign(crypto.GeneratePrivateKey())
	c.AddSignature(v2.Signature, v2.Sender)
	findMessageCertificate(t, v1, c, 2)
	findMessageCertificate(t, v2, c, 2)

	b = make([]byte, c.ByteSize())
	c.MarshallTo(b)
	c = CertificateFromBytes(b)
	findMessageCertificate(t, v1, c, 2)
	findMessageCertificate(t, v2, c, 2)

	v3 := NewSilenceMessage(e, 27)
	v3.Sign(crypto.GeneratePrivateKey())
	c.AddSignature(v3.Signature, v3.Sender)
	findMessageCertificate(t, v1, c, 3)
	findMessageCertificate(t, v2, c, 3)
	findMessageCertificate(t, v3, c, 3)

	b = make([]byte, c.ByteSize())
	c.MarshallTo(b)
	c = CertificateFromBytes(b)
	findMessageCertificate(t, v1, c, 3)
	findMessageCertificate(t, v2, c, 3)
	findMessageCertificate(t, v3, c, 3)
}

func TestSignatureGetCryptoSignatures(t *testing.T) {
	keys := []crypto.PrivateKey{crypto.GeneratePrivateKey(), crypto.GeneratePrivateKey()}
	e := int64(3)
	b := NewBlock(testRandValue(1024), nil)
	votes := []*Message{NewVoteMessage(e, b.BlockID(), b.Height, 0), NewVoteMessage(e, b.BlockID(), b.Height, 1)}
	bc := NewBlockCertificate(e, b.BlockID(), b.Height)
	for i, v := range votes {
		v.Sign(keys[i])
		if v.VerifySignature(keys[i].PubKey()) == false {
			t.Error("Vote is not properly signed!")
		}
		bc.AddSignature(v.Signature, v.Sender)
	}
	for _, sig := range bc.GetCryptoSignatures() {
		if sig.ID != votes[sig.ID].Sender {
			t.Errorf("Expected ID %v got %v\n", sig.ID, votes[sig.ID].Sender)
		}
		if !bytes.Equal(sig.Payload, votes[sig.ID].Payload()) {
			t.Errorf("Expected payload %v got %v\n", sig.Payload, votes[sig.ID].Payload())
		}
		if !bytes.Equal(sig.Signature, votes[sig.ID].Signature) {
			t.Errorf("Expected signature %v got %v\n", sig.Signature, votes[sig.ID].Signature)
		}
	}
	quitEpoch := NewQuitEpochMessage(e, bc)
	marshalled := quitEpoch.Marshall()
	bcc := MessageFromBytes(marshalled).Certificate
	assertCertificateEquals(t, bc, bcc)

	for _, sig := range bcc.GetCryptoSignatures() {
		if sig.ID != votes[sig.ID].Sender {
			t.Errorf("Expected ID %v got %v\n", sig.ID, votes[sig.ID].Sender)
		}
		if !bytes.Equal(sig.Payload, votes[sig.ID].Payload()) {
			t.Errorf("Expected payload %v got %v\n", sig.Payload, votes[sig.ID].Payload())
		}
		if !bytes.Equal(sig.Signature, votes[sig.ID].Signature) {
			t.Errorf("Expected signature %v got %v\n", sig.Signature, votes[sig.ID].Signature)
		}
	}
}
