package mempool

import (
	"bytes"
	"crypto/rand"
	"testing"

	"dslab.inf.usi.ch/tendermint/types"
)

func randomTx(size int) types.Value {
	value := make(types.Value, size)
	rand.Read(value)
	return value
}

func TestMempoolAddGetValue(t *testing.T) {
	m := NewMempool(DefaultConfig())

	// Empty mempool
	v := m.GetValue()
	b := types.ParseBlock(v)
	if len(b.Values) != 0 {
		t.Error("Expected empty block, got", len(b.Values), "values")
	}

	// Add tx, retrieve it via GetValue
	tx := randomTx(32)
	m.Add(tx)
	v = m.GetValue()
	b = types.ParseBlock(v)
	if len(b.Values) != 1 || !bytes.Equal(b.Values[0], tx) {
		t.Error("Expected block with single value", tx, "got:", b.Values)
	}

	// GetValue should not return the same tx again
	v = m.GetValue()
	b = types.ParseBlock(v)
	if len(b.Values) != 0 {
		t.Error("Expected empty block, got values:", b.Values)
	}

	// Add another tx, retrieve it via GetValue
	tx2 := randomTx(32)
	m.Add(tx2)
	v = m.GetValue()
	b = types.ParseBlock(v)
	if len(b.Values) != 1 || !bytes.Equal(b.Values[0], tx2) {
		t.Error("Expected block with single value", tx2, "got:", b.Values)
	}

	// Add the same txs again, they should not be added
	v = m.GetValue()
	b = types.ParseBlock(v)
	m.Add(tx)
	if len(b.Values) != 0 {
		t.Error("Expected empty block, got values:", b.Values)
	}

	v = m.GetValue()
	b = types.ParseBlock(v)
	m.Add(tx2)
	if len(b.Values) != 0 {
		t.Error("Expected empty block, got values:", b.Values)
	}
}

func TestMempoolDecideGetValue(t *testing.T) {
	m := NewMempool(DefaultConfig())

	// Decide a tx not present in the mempool
	tx := randomTx(32)
	bb := new(types.Block)
	bb.Add(tx)
	m.Decide(bb.ToValue())

	// Add the same tx to the mempool, should not be added
	m.Add(tx)
	v := m.GetValue()
	b := types.ParseBlock(v)
	if len(b.Values) != 0 {
		t.Error("Expected empty block, got", len(b.Values), "values")
	}

	// Add a tx to the mempool and decide it, should not be returned then
	tx2 := randomTx(32)
	m.Add(tx2)
	bb = new(types.Block)
	bb.Add(tx2)
	m.Decide(bb.ToValue())

	v = m.GetValue()
	b = types.ParseBlock(v)
	if len(b.Values) != 0 {
		t.Error("Expected empty block, got", len(b.Values), "values")
	}
}
