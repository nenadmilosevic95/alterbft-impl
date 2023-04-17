package consensus

import (
	"fmt"
	"testing"
)

func TestAddBlockWithNoCommit(t *testing.T) {
	blockchain := NewBlockchain(5)
	// Tests A: No commited block!
	// Test A1: Add block with height > MIN_HEIGHT to an empty blockchain
	b0 := NewBlock(testRandValue(1024), nil)
	b1 := NewBlock(testRandValue(1024), b0)
	if blockchain.AddBlock(b1) == false {
		t.Errorf("AddBlock returns false, expected true")
	}
	bb1 := blockchain.getBlock(b1.Height, b1.BlockID())
	if !bb1.Equal(b1) {
		t.Errorf("Expected %v, got %v", b1, bb1)
	}
	// Try to add the same block again
	if blockchain.AddBlock(b1) == true {
		t.Errorf("AddBlock returns true, expected false")
	}
	// Test A2: Add block with height == MIN_HEIGHT to an empty blockchain
	if blockchain.AddBlock(b0) == false {
		t.Errorf("AddBlock returns false, expected true")
	}
	bb0 := blockchain.getBlock(b0.Height, b0.BlockID())
	if !bb0.Equal(b0) {
		t.Errorf("Expected %v, got %v", b0, bb0)
	}
	if bb0.prevBlock != nil {
		t.Errorf("Expected prevBlock %v, got %v", nil, bb0.prevBlock)
	}
	if !bb1.prevBlock.Equal(bb0) {
		t.Errorf("Expected prevBlock %v, got %v", bb0, bb1.prevBlock)
	}
	// Test A3: Add block with predecessor in the blockchain
	b2 := NewBlock(testRandValue(1024), b1)
	if blockchain.AddBlock(b2) == false {
		t.Errorf("AddBlock returns false, expected true")
	}
	bb2 := blockchain.getBlock(b2.Height, b2.BlockID())
	if !bb2.Equal(b2) {
		t.Errorf("Expected %v, got %v", b1, bb1)
	}
	if !bb2.prevBlock.Equal(bb1) {
		t.Errorf("Expected prevBlock %v, got %v", bb1, bb2.prevBlock)
	}
	// Test A4: Add block with no predecessor in the blockchain
	b3 := NewBlock(testRandValue(1024), b2)
	b4 := NewBlock(testRandValue(1024), b3)
	if blockchain.AddBlock(b4) == false {
		t.Errorf("AddBlock returns false, expected true")
	}
	bb4 := blockchain.getBlock(b4.Height, b4.BlockID())
	if !bb4.Equal(b4) {
		t.Errorf("Expected %v, got %v", b4, bb4)
	}
	if !bb4.prevBlock.Equal(nil) {
		t.Errorf("Expected prevBlock %v, got %v", nil, bb4.prevBlock)
	}
	// Test A5: Add block in the already full blockchain
	b5 := NewBlock(testRandValue(1024), b4)
	b6 := NewBlock(testRandValue(1024), b5)
	if blockchain.AddBlock(b6) == true {
		t.Errorf("AddBlock returns true, expected false")
	}
	bb6 := blockchain.getBlock(b6.Height, b6.BlockID())
	if !bb6.Equal(nil) {
		t.Errorf("Expected %v, got %v", nil, bb6)
	}

}

func TestAddBlockWithCommit(t *testing.T) {
	blockchain := NewBlockchain(5)
	// Tests B: Tests with last commited block != nil
	b0 := NewBlock(testRandValue(1024), nil)
	b1 := NewBlock(testRandValue(1024), b0)
	b2 := NewBlock(testRandValue(1024), b1)
	blockchain.AddBlock(b1)
	blockchain.AddBlock(b0)
	blockchain.AddBlock(b2)
	blockchain.Commit(b2)
	b3 := NewBlock(testRandValue(1024), b2)
	// Test 6: Add block where predecessor is last commited block.
	if blockchain.AddBlock(b3) == false {
		t.Errorf("AddBlock returns false, expected true")
	}
	bb3 := blockchain.getBlock(b3.Height, b3.BlockID())
	if !bb3.Equal(b3) {
		t.Errorf("Expected %v, got %v", b3, bb3)
	}
	if !bb3.prevBlock.Equal(b2) {
		t.Errorf("Expected prevBlock %v, got %v", b2, bb3.prevBlock)
	}
	// Test 7: Add block where predecessor is not last commited block.
	b4 := NewBlock(testRandValue(1024), b3)
	if blockchain.AddBlock(b4) == false {
		t.Errorf("AddBlock returns false, expected true")
	}
	bb4 := blockchain.getBlock(b4.Height, b4.BlockID())
	if !bb4.Equal(b4) {
		t.Errorf("Expected %v, got %v", b4, bb4)
	}
	if !bb4.prevBlock.Equal(b3) {
		t.Errorf("Expected prevBlock %v, got %v", b3, bb4.prevBlock)
	}
	// Test 8: Add block where predecessor is not in the blockchain.
	b5 := NewBlock(testRandValue(1024), b4)
	b6 := NewBlock(testRandValue(1024), b5)
	if blockchain.AddBlock(b6) == false {
		t.Errorf("AddBlock returns false, expected true")
	}
	bb6 := blockchain.getBlock(b6.Height, b6.BlockID())
	if !bb6.Equal(b6) {
		t.Errorf("Expected %v, got %v", b6, bb6)
	}
	// Test 9: Add block in the already full blockchain.
	b7 := NewBlock(testRandValue(1024), b6)
	if blockchain.AddBlock(b7) == true {
		t.Errorf("AddBlock returns true, expected false")
	}
	bb7 := blockchain.getBlock(b7.Height, b7.BlockID())
	if !bb7.Equal(nil) {
		t.Errorf("Expected %v, got %v", nil, bb7)
	}
}

func TestCommitBlock(t *testing.T) {
	blockchain := NewBlockchain(3)
	// Test 1: Commit initial block of the blockchain.
	b0 := NewBlock(testRandValue(1024), nil)
	blockchain.AddBlock(b0)
	blocks := blockchain.Commit(b0)
	if len(blocks) != 1 || !blocks[0].Equal(b0) {
		t.Errorf("Commit returns %v, expected %v", blocks, b0)
	}

	// Test 2: Commit block extending last commited block.
	b1 := NewBlock(testRandValue(1024), b0)
	blockchain.AddBlock(b1)
	blocks = blockchain.Commit(b1)
	if len(blocks) != 1 || !blocks[0].Equal(b1) {
		t.Errorf("Commit returns %v, expected %v", blocks, b1)
	}

	// Test 3: Commit block not extending directly last commited block.
	b2 := NewBlock(testRandValue(1024), b1)
	b3 := NewBlock(testRandValue(1024), b2)
	blockchain.AddBlock(b2)
	blockchain.AddBlock(b3)
	blocks = blockchain.Commit(b3)
	if len(blocks) != 2 || !blocks[0].Equal(b3) || !blocks[1].Equal(b2) {
		t.Errorf("Commit returns %v, expected %v and %v", blocks, b3, b2)
	}
}

func TestExtend(t *testing.T) {
	blockchain := NewBlockchain(10)
	// Test 1
	b0 := NewBlock(testRandValue(1024), nil)
	bb0 := NewBlock(testRandValue(1024), nil)
	blockchain.AddBlock(b0)
	blockchain.AddBlock(bb0)
	if blockchain.ExtendValidChain(b0) == false {
		t.Error("ExtendValidChain() returned false expected true!")
	}
	if blockchain.ExtendValidChain(bb0) == false {
		t.Error("ExtendValidChain() returned false expected true!")
	}
	b1 := NewBlock(testRandValue(1024), b0)
	blockchain.AddBlock(b1)
	if blockchain.ExtendValidChain(b1) == false {
		t.Error("ExtendValidChain() returned false expected true!")
	}
	b2 := NewBlock(testRandValue(1024), b1)
	b3 := NewBlock(testRandValue(1024), b2)
	blockchain.AddBlock(b3)
	if blockchain.ExtendValidChain(b3) == true {
		t.Error("ExtendValidChain() returned true expected false!")
	}
	blockchain.AddBlock(b2)
	if blockchain.ExtendValidChain(b3) == false {
		t.Error("ExtendValidChain() returned false expected true!")
	}
	bb1 := NewBlock(testRandValue(1024), bb0)
	blockchain.AddBlock(bb1)
	if blockchain.ExtendValidChain(bb1) == false {
		t.Error("ExtendValidChain() returned false expected true!")
	}
	blockchain.Commit(bb1)
	if blockchain.ExtendValidChain(b0) == true {
		t.Error("ExtendValidChain() returned true expected false!")
	}
	if blockchain.ExtendValidChain(b1) == true {
		t.Error("ExtendValidChain() returned true expected false!")
	}
	if blockchain.ExtendValidChain(b2) == true {
		t.Error("ExtendValidChain() returned true expected false!")
	}
	if blockchain.ExtendValidChain(b3) == true {
		t.Error("ExtendValidChain() returned true expected false!")
	}
	bb2 := NewBlock(testRandValue(1024), bb1)
	bb3 := NewBlock(testRandValue(1024), bb2)
	blockchain.AddBlock(bb3)
	if blockchain.ExtendValidChain(bb3) == true {
		t.Error("ExtendValidChain() returned true expected false!")
	}
	blockchain.AddBlock(bb2)
	if blockchain.ExtendValidChain(bb2) == false {
		t.Error("ExtendValidChain() returned false expected true!")
	}
	if blockchain.ExtendValidChain(bb3) == false {
		t.Error("ExtendValidChain() returned false expected true!")
	}

}

func (b *Blockchain) print() string {
	s := fmt.Sprintf("Last commited: %v\n", b.LastCommited.BlockID())
	for i := 0; i < b.Size; i++ {
		s = fmt.Sprintf("%vHeight: %v\n", s, i)
		hd := b.getHeightData(int64(i))
		s = fmt.Sprintf("%vCandidates:\n %v", s, hd.candidates)
	}
	return s
}
