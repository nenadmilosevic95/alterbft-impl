package consensus

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// Block is a block abstraction
type Block struct {
	Height      int64   // Block height
	Value       []byte  // Block value
	PrevBlockID BlockID // Hash of the previous block

	// Help fields, these fields we don't marshall
	marshalled []byte
	// Set by the blockchain when we receive a proposal from the network.
	// Used only for blockchain chaining.
	prevBlock *Block
	// Set in order not to calcualte hash of a block each time
	blockID BlockID
}

func (b *Block) String() string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("Height: %v\nValue: %v\nPrevBlock:%v\n", b.Height, b.Value, b.PrevBlockID)
}

// NewBlock returns a pointer to an instance of block.
func NewBlock(v []byte, prevBlock *Block) *Block {
	h := int64(MIN_HEIGHT)
	var prevBlockID BlockID = nil
	if prevBlock != nil {
		h = prevBlock.Height + 1
		prevBlockID = prevBlock.BlockID()
	}
	return &Block{
		Height:      h,
		Value:       v,
		PrevBlockID: prevBlockID,
		// we don't set here prevBlock because anyway we are not going to marshall it
		// so this info will be lost.
	}
}

// Equal tests for equality among two blocks.
func (b *Block) Equal(bb *Block) bool {
	if b == nil || bb == nil {
		return b == bb
	}
	return (b.Height == bb.Height &&
		bytes.Equal(b.Value, bb.Value) &&
		b.PrevBlockID.Equal(bb.PrevBlockID))
}

// Extend checks whether b extend bb.
func (b *Block) Extend(bb *Block) bool {
	if b == nil {
		return false
	}
	if bb == nil {
		return b.Height == MIN_HEIGHT
	}
	if b.Height == bb.Height+1 && b.PrevBlockID.Equal(bb.BlockID()) {
		return true
	}
	if b.Height == bb.Height && b.BlockID().Equal(bb.BlockID()) {
		return true
	}
	return false
}

// BlockID returns a hash of the full block.
func (b *Block) BlockID() BlockID {
	if b.blockID == nil {
		if b.marshalled == nil {
			b.Marshall()
		}
		hash := sha256.Sum256(b.marshalled)
		b.blockID = hash[:]
	}
	return b.blockID
}

// ByteSize returns block size in bytes.
func (b *Block) ByteSize() int {
	if b.PrevBlockID == nil {
		return len(b.Value) + 8
	} else {
		return len(b.Value) + BlockIDSize + 8
	}
}

// Marshall returns marshalled version of a block.
func (b *Block) Marshall() []byte {
	if b.marshalled == nil {
		b.marshalled = make([]byte, b.ByteSize())
		b.MarshallTo(b.marshalled)
	}
	return b.marshalled
}

// MarshallTo marshall block to the buffer.
func (b *Block) MarshallTo(buffer []byte) int {
	encoding.PutUint64(buffer[0:], uint64(b.Height))
	pos := 8
	if b.PrevBlockID != nil {
		b.PrevBlockID.MarshallTo(buffer[pos:])
		pos += BlockIDSize
	}
	copy(buffer[pos:], b.Value)
	return pos + len(b.Value)
}

// BlockFromBytes unmarshall block from a buffer.
func BlockFromBytes(buffer []byte) *Block {
	height := int64(encoding.Uint64(buffer[0:]))
	var prevBlockID BlockID = nil
	pos := 8
	if height > MIN_HEIGHT { // it has prevBlockID
		prevBlockID = BlockIDFromBytes(buffer[pos:])
		pos += BlockIDSize
	}
	value := buffer[pos:]
	return &Block{
		Height:      height,
		Value:       value,
		PrevBlockID: prevBlockID,

		marshalled: buffer,
	}
}

const BlockIDSize = sha256.Size

// BlockID represents a hash of a block.
type BlockID []byte

// Equal returns whether a value equals another.
func (b BlockID) Equal(bb BlockID) bool {
	return bytes.Equal(b, bb)
}

// ByteSize returns blockID size in bytes.
func (b BlockID) ByteSize() int {
	return BlockIDSize
}

// MarshallTo marshall blockID to the buffer.
func (b BlockID) MarshallTo(buffer []byte) int {
	copy(buffer[:BlockIDSize], b)
	return BlockIDSize
}

// BlockIDFromBytes unmarshall blockID from a buffer.
func BlockIDFromBytes(buffer []byte) BlockID {
	return buffer[:BlockIDSize]
}
