package crypto

import "testing"

func TestGeneratedPrivateKey(t *testing.T) {
	priv1 := GeneratePrivateKey()
	priv2 := GeneratePrivateKey()
	if priv1 == nil || priv2 == nil {
		t.Fatal("Unexpected nil private keys", priv1, priv2)
	}
	if priv1.Equals(priv2) {
		t.Error("Unexpected generated private keys to be equal", priv1, priv2)
	}

	pub1 := priv1.PubKey()
	pub2 := priv2.PubKey()
	if pub1 == nil || pub2 == nil {
		t.Fatal("Unexpected nil public keys", pub1, pub2)
	}

	m := []byte("a message")
	sig1, err := priv1.Sign(m)
	if sig1 == nil || err != nil {
		t.Error("Unexpected nil signature or error", sig1, err)
	}
	if len(sig1) != SignatureSize {
		t.Error("Produced signature size expected", SignatureSize, "got", len(sig1))
	}
	sig2, err := priv2.Sign(m)
	if sig2 == nil || err != nil {
		t.Error("Unexpected nil signature or error", sig2, err)
	}
	if len(sig2) != SignatureSize {
		t.Error("Produced signature size expected", SignatureSize, "got", len(sig2))
	}

	if !pub1.VerifySignature(m, sig1) {
		t.Error("Failed to verify signature produced by the key")
	}
	if !pub2.VerifySignature(m, sig2) {
		t.Error("Failed to verify signature produced by the key")
	}
	if pub1.VerifySignature(m, sig2) {
		t.Error("Unexpected to verify signature produced by other key")
	}
	if pub2.VerifySignature(m, sig1) {
		t.Error("Unexpected to verify signature produced by other key")
	}
}

func TestPrivateKeyFromSecret(t *testing.T) {
	s := []byte("my secret")

	priv1 := GeneratePrivateKeyFromSecret(s)
	priv2 := GeneratePrivateKeyFromSecret(s)
	if priv1 == nil || priv2 == nil {
		t.Fatal("Unexpected nil private keys", priv1, priv2, s)
	}
	if !priv1.Equals(priv2) {
		t.Error("Expected private keys form the same secret to be equal",
			priv1, priv2, s)
	}

	pub1 := priv1.PubKey()
	pub2 := priv2.PubKey()
	if pub1 == nil || pub2 == nil {
		t.Fatal("Unexpected nil public keys", pub1, pub2)
	}

	m := []byte("a message")
	sig1, err := priv1.Sign(m)
	if sig1 == nil || err != nil {
		t.Error("Unexpected nil signature or error", sig1, err)
	}
	if len(sig1) != SignatureSize {
		t.Error("Produced signature size expected", SignatureSize, "got", len(sig1))
	}
	sig2, err := priv2.Sign(m)
	if sig2 == nil || err != nil {
		t.Error("Unexpected nil signature or error", sig2, err)
	}
	if len(sig2) != SignatureSize {
		t.Error("Produced signature size expected", SignatureSize, "got", len(sig2))
	}

	if !pub2.VerifySignature(m, sig1) {
		t.Error("Failed to verify signature produced by the key")
	}
	if !pub1.VerifySignature(m, sig2) {
		t.Error("Failed to verify signature produced by the key")
	}
}
