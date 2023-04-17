package types

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func randomValue(size int) Value {
	value := make(Value, size)
	rand.Read(value)
	return value
}

func TestBlockEmptyValue(t *testing.T) {
	values_num := 0
	e_size := 4

	b := &Block{}
	if len(b.Values) != values_num {
		t.Error("Block values: expected", values_num, "got", len(b.Values))
	}

	size := b.ByteSize()
	if size != e_size {
		t.Error("Block byte size: expected", e_size, "got", size)
	}

	v := b.ToValue()
	if len(v) != size {
		t.Error("len(v): expected", size, "got", len(v))
	}

	bb := ParseBlock(v)
	if bb == nil {
		t.Fatal("BlockFromValue() returned", bb)
	}
	if bb.ByteSize() != size {
		t.Error("Block byte size: expected", size, "got", bb.ByteSize())
	}
	if len(bb.Values) != values_num {
		t.Error("Block values: expected", values_num, "got", len(bb.Values))
	}
}

func TestBlockOneValueValue(t *testing.T) {
	values_num := 1
	e_size := 4

	b := &Block{}

	v := randomValue(16)
	if bytes.Equal(v, make([]byte, 16)) {
		t.Fatal("Not random value:", v)
	}
	b.Add(v)
	if !bytes.Equal(b.Values[0], v) {
		t.Error("Value", 0, "expected", v, "got", b.Values[0])
	}

	e_size += 4 + len(v)
	if len(b.Values) != values_num {
		t.Error("Block values: expected", values_num, "got", len(b.Values))
	}

	size := b.ByteSize()
	if size != e_size {
		t.Error("Block byte size: expected", e_size, "got", size)
	}

	value := b.ToValue()
	if len(value) != size {
		t.Error("len(v): expected", size, "got", len(value))
	}

	bb := ParseBlock(value)
	if bb == nil {
		t.Fatal("BlockFromValue() returned", bb)
	}
	if bb.ByteSize() != size {
		t.Error("Block byte size: expected", size, "got", bb.ByteSize())
	}
	if len(bb.Values) != values_num {
		t.Error("Block values: expected", values_num, "got", len(bb.Values))
	}

	for i, v := range bb.Values {
		evalue := b.Values[i]
		if !bytes.Equal(v, evalue) {
			t.Error("Value", i, "expected", evalue, "got", v)
		}
	}
}

func TestBlockTwoValuesValue(t *testing.T) {
	values_num := 2
	e_size := 4

	b := &Block{}

	v := randomValue(16)
	if bytes.Equal(v, make([]byte, 16)) {
		t.Fatal("Not random value:", v)
	}
	b.Add(v)
	if !bytes.Equal(b.Values[0], v) {
		t.Error("Value", 0, "expected", v, "got", b.Values[0])
	}
	e_size += 4 + len(v)

	v2 := randomValue(16)
	if bytes.Equal(v2, make([]byte, 16)) {
		t.Fatal("Not random value:", v2)
	}
	if bytes.Equal(v2, v) {
		t.Fatal("Not random value:", v2, v)
	}
	b.Add(v2)
	if !bytes.Equal(b.Values[1], v2) {
		t.Error("Value", 1, "expected", v2, "got", b.Values[1])
	}
	e_size += 4 + len(v2)

	if len(b.Values) != values_num {
		t.Error("Block values: expected", values_num, "got", len(b.Values))
	}

	size := b.ByteSize()
	if size != e_size {
		t.Error("Block byte size: expected", e_size, "got", size)
	}

	value := b.ToValue()
	if len(value) != size {
		t.Error("len(v): expected", size, "got", len(value))
	}

	bb := ParseBlock(value)
	if bb == nil {
		t.Fatal("BlockFromValue() returned", bb)
	}
	if bb.ByteSize() != size {
		t.Error("Block byte size: expected", size, "got", bb.ByteSize())
	}
	if len(bb.Values) != values_num {
		t.Error("Block values: expected", values_num, "got", len(bb.Values))
	}

	for i, value := range bb.Values {
		evalue := b.Values[i]
		if !bytes.Equal(value, evalue) {
			t.Error("Value", i, "expected", evalue, "got", value)
		}
	}
}
