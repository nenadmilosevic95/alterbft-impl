package tendermint

/*
func TestVerifier(t *testing.T) {
	privKeys := make([]crypto.PrivateKey, 2)
	pubKeys := make([]crypto.PublicKey, 2)
	for i := range privKeys {
		privKeys[i] = crypto.GeneratePrivateKey()
		pubKeys[i] = privKeys[i].PubKey()
	}
	input := make(chan net.Message, 1)
	v := NewVerifier(pubKeys, input, nil)

	// 1. Produce a consensus message
	// 2. Marshall the message and add it to the verifier's input queue
	// 3. The original message should be read from the verifier's output queue

	m1 := &consensus.Message{
		Type:   consensus.VOTE,
		Epoch:  17,
		Sender: 0,
		//BlockID: consensus.NewBlockID(),
	}
	m1.Sign(privKeys[0])
	sig1 := m1.GetCryptoSignatures()[0]
	t.Log("Sig", 0, sig1, pubKeys[sig1.ID].VerifySignature(sig1.Payload, sig1.Signature))
	m2 := &consensus.Message{
		Type:   consensus.VOTE,
		Epoch:  17,
		Sender: 1,
		//BlockID: consensus.NewBlockID(),
	}
	m2.Sign(privKeys[1])
	sig2 := m2.GetCryptoSignatures()[0]
	t.Log("Sig", 1, sig2, pubKeys[sig2.ID].VerifySignature(sig2.Payload, sig2.Signature))
	var ms []*consensus.Message
	ms = append(ms, m1)
	ms = append(ms, m2)
	var cert *consensus.Certificate
	//cert := consensus.NewBlockCertificate(ms)
	mc := &consensus.Message{
		Type:        consensus.QUIT_EPOCH,
		Sender:      0,
		Certificate: cert,
	}
	mc.Sign(privKeys[0])
	b := mc.Marshall()

	mc = consensus.MessageFromBytes(b)
	sigs := mc.GetCryptoSignatures()
	t.Log(mc)
	for i, sig := range sigs {
		t.Log("Sig", i, sig, pubKeys[sig.ID].VerifySignature(sig.Payload, sig.Signature))
	}

	input <- m1.Marshall()

	select {
	case m := <-v.Output():
		if !m.Equal(m1) {
			t.Error("Output message", m, "differs from", m1)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Not output generated in 100ms")
	}

	input <- mc.Marshall()

	select {
	case m := <-v.Output():
		if !m.Equal(mc) {
			t.Error("Output message", m, "differs from", m1)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Not output generated in 100ms")
	}

	// Produce a consensus message with a valid certificate

	// Produce a consensus message and trick its signatures

	// Produce a consensus message with a certificate and trick one of the signatures

	// FIXME: cache testing??
}
*/
