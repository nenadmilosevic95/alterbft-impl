package tendermint

// Stats for a process.
type Stats struct {
	Instances  [3]int  // Started, Decided, Delivered
	Messages   [11]int // PROPOSAL, PREVOTE, PRECOMMIT, VALUE
	Deliveries [2]int  // Blocks, Transactions
}

func NewStats() *Stats {
	return &Stats{}
}

func (s *Stats) InstanceStarted() {
	s.Instances[0] += 1
}

func (s *Stats) InstanceDecided() {
	s.Instances[1] += 1
}

func (s *Stats) InstanceDelivered(txs int) {
	s.Instances[2] += 1
	s.Deliveries[0] += 1
	if txs > 0 {
		s.Deliveries[1] += txs
	}
}

func (s *Stats) MessageReceived(mtype int16) {
	s.Messages[mtype] += 1
}
