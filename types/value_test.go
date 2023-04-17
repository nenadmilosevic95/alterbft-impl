package types

import (
	"testing"
)

func TestValueEqual(t *testing.T) {
	v1 := Value("abc")
	if v1 == nil {
		t.Error("Unexpected nil value", v1)
	}
	if !v1.Equal(v1) {
		t.Error("Value is expected to be equal to itself", v1)
	}

	v2 := Value("ab")
	if v2 == nil {
		t.Error("Unexpected nil value", v2)
	}
	if !v2.Equal(v2) {
		t.Error("Value is expected to be equal to itself", v2)
	}

	v3 := Value("abc")
	if v3 == nil {
		t.Error("Unexpected nil value", v3)
	}
	if !v3.Equal(v3) {
		t.Error("Value is expected to be equal to itself", v3)
	}

	if v2.Equal(v1) || v1.Equal(v2) {
		t.Error("Values 1 and 2 expected to not be equal", v1, v2)
	}
	if !v1.Equal(v3) || !v3.Equal(v1) {
		t.Error("Values 1 and 3 expected to be equal", v1, v3)
	}
	if v2.Equal(v3) || v3.Equal(v2) {
		t.Error("Values 2 and 3 expected to not be equal", v2, v3)
	}
}

func TestValueID(t *testing.T) {
	v := Value("abc")
	id := v.ID()
	if id == nil {
		t.Error("Unexpected nil ID for value", v, id)
	}
	if v.Equal(id) || id.Equal(v) {
		t.Error("Unexpected value to equal its ID", v, id)
	}
}

func TestValueIDEqual(t *testing.T) {
	v1 := Value("abc")
	v2 := Value("ab")
	v3 := Value("abc")
	if v1.ID().Equal(v2.ID()) || v2.ID().Equal(v1.ID()) {
		t.Error("Values 1 and 2 differ", v1, v2, "but their IDs equal", v1.ID(), v2.ID())
	}
	if !v1.ID().Equal(v3.ID()) || !v3.ID().Equal(v1.ID()) {
		t.Error("Values 1 and 3 equal", v1, v3, "but their IDs differ", v1.ID(), v3.ID())
	}
	if v2.ID().Equal(v3.ID()) || v3.ID().Equal(v2.ID()) {
		t.Error("Values 2 and 3 differ", v2, v3, "but their IDs equal", v2.ID(), v3.ID())
	}
}

func TestValueNil(t *testing.T) {
	v := Value(nil)
	id := v.ID()
	if v != nil {
		t.Error("Expected nil Value got", v)
	}
	if id != nil {
		t.Error("Expected nil ID for nil Value got", id)
	}
}
