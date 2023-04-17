package consensus

import "fmt"

type heightData struct {
	height int64
	//block      *Block
	candidates []*Block
}

func NewHeightData(height int64) *heightData {
	return &heightData{
		height: height,
	}
}

func (hd *heightData) addCandidate(block *Block) {
	hd.candidates = append(hd.candidates, block)
}

func (hd *heightData) getCandidate(blockID BlockID) *Block {
	if hd == nil {
		return nil
	}
	for _, b := range hd.candidates {
		if b.BlockID().Equal(blockID) {
			return b
		}
	}
	return nil
}

type Blockchain struct {
	Size         int
	LastCommited *Block
	Chain        []*heightData
}

func NewBlockchain(size int) *Blockchain {
	return &Blockchain{
		Size:         size,
		LastCommited: nil,
		Chain:        make([]*heightData, size),
	}
}

// AddBlock add block to the blockchain, only condition is that it needs to have
// height higher than commited block.
func (b *Blockchain) AddBlock(block *Block) bool {
	if b.getBlock(block.Height, block.BlockID()) != nil {
		return false
	}
	hasValidHeight := (b.LastCommited == nil && block.Height < int64(b.Size)) ||
		(b.LastCommited != nil &&
			block.Height > b.LastCommited.Height &&
			block.Height < b.LastCommited.Height+int64(b.Size))

	if !hasValidHeight {
		return false
	}
	if b.LastCommited != nil {
		isExtendingLastCommitedBlock := block.Extend(b.LastCommited)
		isPotentialNewBlock := block.Height > b.LastCommited.Height+1
		if !isExtendingLastCommitedBlock && !isPotentialNewBlock {
			return false
		}
	}
	var prevBlock *Block
	if block.Height == MIN_HEIGHT {
		prevBlock = nil
	} else {
		prevBlock = b.getBlock(block.Height-1, block.PrevBlockID)
	}
	block.prevBlock = prevBlock
	b.addBlock(block)
	b.updateSuccessor(block)
	return true
}

func (b *Blockchain) updateSuccessor(block *Block) {
	hd := b.getHeightData(block.Height + 1)
	if hd != nil {
		for _, bb := range hd.candidates {
			if bb.PrevBlockID.Equal(block.BlockID()) {
				bb.prevBlock = block
			}
		}
	}
}

// This a Commit version for byzantine attacks.
func (b *Blockchain) CommitByzantine(block *Block) []*Block {
	var blocks []*Block
	blocks = append(blocks, block)
	bb := block.prevBlock
	lastCommitedHeight := int64(0)
	if b.LastCommited != nil {
		lastCommitedHeight = b.LastCommited.Height
	}
	for bb != nil && bb.Height > lastCommitedHeight {
		blocks = append(blocks, bb)
		bb = bb.prevBlock
	}
	b.LastCommited = block
	return blocks
}

// GetBlock returns full block.
// Contract: used only from AddBlock and in tests, so all checks are previously made.
func (b *Blockchain) getBlock(h int64, blockID BlockID) *Block {
	if b.LastCommited != nil && b.LastCommited.BlockID().Equal(blockID) {
		return b.LastCommited
	}
	hd := b.getHeightData(h)
	return hd.getCandidate(blockID)
}

// addBlock adds a block to the list of candidates.
// Contract: used only from AddBlock, so all checks are previously made.
func (b *Blockchain) addBlock(block *Block) {
	hd := b.getHeightData(block.Height)
	hd.addCandidate(block)
}

// Commit is called when consensus commit on a block,
// it should delete all unnecessary blocks and return part of
// the chain that is commited with this block from this block
// to the lastly commited block.
// Contract: All blocks that should be returned/commited are in the blockchain already!
func (b *Blockchain) Commit(block *Block) []*Block {
	var blocks []*Block
	blocks = append(blocks, block)
	b.resetHeightData(block.Height)
	bb := block.prevBlock
	for !bb.Equal(b.LastCommited) {
		blocks = append(blocks, bb)
		b.resetHeightData(bb.Height)
		bb = bb.prevBlock
	}
	block.prevBlock = nil
	b.LastCommited = block
	return blocks
}

// ExtendValidChain returns if block extend last commited block.
func (b *Blockchain) ExtendValidChain(block *Block) bool {
	tmp := block
	for tmp.prevBlock != nil {
		tmp = tmp.prevBlock
	}
	if b.LastCommited == nil && tmp.Height != MIN_HEIGHT {
		return false
	}
	if b.LastCommited != nil && !tmp.Equal(b.LastCommited) {
		return false
	}
	return true
}

func (b *Blockchain) resetHeightData(height int64) {
	index := height % int64(b.Size)
	if b.Chain[index] != nil {
		b.Chain[index].height = -1
		b.Chain[index].candidates = nil
	}
}

func (b *Blockchain) IsEquivocatedBlock(block *Block) bool {
	return false
}

func (b *Blockchain) getHeightData(height int64) *heightData {
	if (b.LastCommited == nil && height >= int64(b.Size)) ||
		(b.LastCommited != nil && height >= b.LastCommited.Height+int64(b.Size)) {
		fmt.Errorf("Blockchain is to small!")
		return nil
	}
	if height < MIN_HEIGHT {
		return nil
	}
	index := height % int64(b.Size)
	if b.Chain[index] == nil || b.Chain[index].height < height {
		b.Chain[index] = NewHeightData(height)
	}
	return b.Chain[index]
}
