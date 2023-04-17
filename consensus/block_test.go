package consensus

import (
	"math/rand"
	"testing"
	"time"
)

func testRandValue(size int) []byte {
	v := make([]byte, size)
	rand.Seed(time.Now().UnixNano())
	rand.Read(v)
	return v
}

func TestBlockEqual(t *testing.T) {
	var b1, b2 *Block = nil, nil
	// Test 1: Both blocks are nil.
	if !b1.Equal(b2) {
		t.Errorf("Equal returns false for \n%v,\n%v", b1, b2)
	}
	// Test 2: First block is nil.
	b2 = NewBlock(testRandValue(1024), nil)
	if b1.Equal(b2) {
		t.Errorf("Equal returns true for \n%v,\n%v", b1, b2)
	}
	// Test 3: Second block is nil.
	b2 = nil
	b1 = NewBlock(testRandValue(1024), nil)
	if b1.Equal(b2) {
		t.Errorf("Equal returns true for \n%v,\n%v", b1, b2)
	}
	// Test 4: Two same blocks with prevBlockID equal nil.
	v := testRandValue(1024)
	b1 = NewBlock(v, nil)
	b2 = NewBlock(v, nil)
	if !b1.Equal(b2) {
		t.Errorf("Equal returns false for: \n%v\n %v", b1, b2)
	}
	// Test 5: Two same blocks with prevBlockID not equal nil.
	b0 := NewBlock(testRandValue(1024), nil)
	v = testRandValue(1024)
	b1 = NewBlock(v, b0)
	b2 = NewBlock(v, b0)
	if !b1.Equal(b2) {
		t.Errorf("Equal returns false for: \n%v\n %v", b1, b2)
	}
	// Test 6: Two blocks with different values.
	b1 = NewBlock(testRandValue(1024), b0)
	b2 = NewBlock(testRandValue(1024), b0)
	if b1.Equal(b2) {
		t.Errorf("Equal returns true for: \n%v\n %v", b1, b2)
	}
	// Test 7: Two blocks with from different heights.
	b1 = NewBlock(testRandValue(1024), b0)
	b2 = NewBlock(testRandValue(1024), b1)
	if b1.Equal(b2) {
		t.Errorf("Equal returns true for: \n%v\n %v", b1, b2)
	}
	// Test 8: Two blocks with different previous block.
	b0_a := NewBlock(testRandValue(1024), nil)
	b0_b := NewBlock(testRandValue(1024), nil)
	v = testRandValue(1024)
	b1 = NewBlock(v, b0_a)
	b2 = NewBlock(v, b0_b)
	if b1.Equal(b2) {
		t.Errorf("Equal returns true for: \n%v\n %v", b1, b2)
	}
}

func TestBlockExtend(t *testing.T) {
	// Test 1: Nil block doesn't extend anything.
	b1 := NewBlock(testRandValue(1024), nil)
	var b2 *Block = nil
	if b2.Extend(b1) {
		t.Errorf("Extend returns true for %v %v", nil, b1)
	}
	// Test 2: The first block in the chain.
	if !b1.Extend(nil) {
		t.Errorf("Extend returns false for %v %v", b1, nil)
	}
	// Test 3: Block extends first block.
	b2 = NewBlock(testRandValue(1024), b1)
	if !b2.Extend(b1) {
		t.Errorf("Extend returns false for %v %v", b2, b1)
	}
	// Test 4: Block extends itself.
	if !b1.Extend(b1) {
		t.Errorf("Extend returns true for %v %v", b1, b1)
	}
	// Test 5: Block with good height but extending different block.
	b1_a := NewBlock(testRandValue(1024), nil)
	b2 = NewBlock(testRandValue(1024), b1_a)
	if b2.Extend(b1) {
		t.Errorf("Extend returns true for %v %v", b2, b1)
	}
}

func TestBlockID(t *testing.T) {

	// Test 1: BlockID() is returning hash and setting block.blockID field.
	b0 := NewBlock(testRandValue(1024), nil)
	blockID := b0.BlockID()
	if !b0.blockID.Equal(blockID) {
		t.Errorf("BlockID() is not returning hash and setting block.blockID field")
	}
	// Test 2: BlockID of two same blocks with no predecessor.
	v := testRandValue(1024)
	b1 := NewBlock(v, nil)
	b2 := NewBlock(v, nil)
	if !b1.BlockID().Equal(b2.BlockID()) {
		t.Errorf("BlockID() is returning different values for the same blocks!")
	}
	// Test 3: BlockID of two same blocks with predecessor.
	b1 = NewBlock(v, b0)
	b2 = NewBlock(v, b0)
	if !b1.BlockID().Equal(b2.BlockID()) {
		t.Errorf("BlockID() is returning different values for the same blocks!")
	}
	// Test 4: BlockID of two blocks with different height.
	if b0.BlockID().Equal(b1.BlockID()) {
		t.Errorf("BlockID() is returning same values for the blocks with different heights!")
	}
	// Test 5: BlockID of two blocks with different value.
	b1 = NewBlock(testRandValue(1024), b0)
	b2 = NewBlock(testRandValue(1024), b0)
	if b1.BlockID().Equal(b2.BlockID()) {
		t.Errorf("BlockID() is returning same values for the blocks with different values!")
	}
	// Test 6: BlockID of two blocks with different previous block.
	b0_a := NewBlock(testRandValue(1024), nil)
	b1 = NewBlock(testRandValue(1024), b0_a)
	b2 = NewBlock(testRandValue(1024), b0)
	if b1.BlockID().Equal(b2.BlockID()) {
		t.Errorf("BlockID() is returning same values for the blocks with different values!")
	}
}

func TestBlockMarshalling(t *testing.T) {
	// Test 1: Marshalling of block with no predecessor.
	b0 := NewBlock(testRandValue(1024), nil)
	b0.Marshall()
	b := BlockFromBytes(b0.marshalled)
	if !b0.Equal(b) {
		t.Errorf("Marshalling of block with no predecessor is not working expected \n%v, returned \n%v\n", b0, b)
	}
	// Test 2: Marshalling of block with predecessor.
	b1 := NewBlock(testRandValue(1024), b0)
	b1.Marshall()
	b = BlockFromBytes(b1.marshalled)
	if !b1.Equal(b) {
		t.Errorf("Marshalling of block with  predecessor is not working expected \n%v, returned \n%v\n", b1, b)
	}
}

func TestBlockIDEqual(t *testing.T) {
	b := NewBlock(testRandValue(1024), nil)
	bb := NewBlock(testRandValue(1024), nil)
	var b1, b2 BlockID = nil, nil
	// Test 1: Both blockIDs are nil.
	if !b1.Equal(b2) {
		t.Errorf("Equal returns false for \n%v,\n%v", b1, b2)
	}
	// Test 2: First blockID is nil.
	b2 = b.BlockID()
	if b1.Equal(b2) {
		t.Errorf("Equal returns true for \n%v,\n%v", b1, b2)
	}
	// Test 3: Second blockID is nil.
	b2 = nil
	b1 = b.BlockID()
	if b1.Equal(b2) {
		t.Errorf("Equal returns true for \n%v,\n%v", b1, b2)
	}
	// Test 4: Two same blockIDs.
	b1 = b.BlockID()
	b2 = b.BlockID()
	if !b1.Equal(b2) {
		t.Errorf("Equal returns false for \n%v,\n%v", b1, b2)
	}
	// Test 5: Two different blockIDs.
	b1 = b.BlockID()
	b2 = bb.BlockID()
	if b1.Equal(b2) {
		t.Errorf("Equal returns true for \n%v,\n%v", b1, b2)
	}
}

func TestBlockIDMarshalling(t *testing.T) {
	// Test 1: Marshalling of blockID.
	block := NewBlock(testRandValue(1024), nil)
	b := block.BlockID()
	buffer := make([]byte, b.ByteSize())
	b.MarshallTo(buffer)
	bb := BlockIDFromBytes(buffer)
	if !b.Equal(bb) {
		t.Errorf("Marshalling of blockID is not working expected \n%v, returned \n%v\n", b, bb)
	}
}
