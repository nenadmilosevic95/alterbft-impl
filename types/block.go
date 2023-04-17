package types

import (
	"encoding/binary"
)

var encoding = binary.LittleEndian

// Block is a batch of values.
type Block struct {
	Values []Value
}

// Add a value to the block.
func (b *Block) Add(value Value) {
	b.Values = append(b.Values, value)
}

// ByteSize returns the total size of the block.
func (b *Block) ByteSize() (size int) {
	for i := range b.Values {
		size += 4 + len(b.Values[i])
	}
	return 4 + size
}

// ToValue produces a value to propose for consensus.
func (b *Block) ToValue() Value {
	value := make([]byte, b.ByteSize())
	encoding.PutUint32(value, uint32(len(b.Values))) // num values
	index := 4                                       // len(Uint32)
	for _, v := range b.Values {
		encoding.PutUint32(value[index:], uint32(len(v)))
		index += 4 // len(Uint32)

		copy(value[index:], v)
		index += len(v)
	}
	return value
}

// ParseBlock builds a block from a value.
func ParseBlock(value Value) *Block {
	valuesNum := int(encoding.Uint32(value))
	index := 4 // len(Uint32)

	block := &Block{}
	for i := 0; i < valuesNum; i++ {
		size := int(encoding.Uint32(value[index:]))
		index += 4 // len(Uint32)

		v := make(Value, size)
		copy(v, value[index:])
		index += size
		block.Add(v)
	}
	return block
}
