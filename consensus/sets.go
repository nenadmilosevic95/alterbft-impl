package consensus

// ProposalSet stores proposals for an epoch of consensus.
type ProposalSet struct {
	proposals []*Message
}

// NewProposalSet creates a ProposalSet.
func NewProposalSet() *ProposalSet {
	return &ProposalSet{}
}

// Add a proposal to the set
// Contract: no dupplication, check is made by a caller before calling this method!
func (ps *ProposalSet) Add(m *Message) {
	ps.proposals = append(ps.proposals, m)
}

func (ps *ProposalSet) Get(blockID BlockID) *Message {
	index := ps.getIndex(blockID)
	if index != -1 {
		return ps.proposals[index]
	}
	return nil
}

// Has check whether proposal with this blockID is present.
func (ps *ProposalSet) Has(blockID BlockID) bool {
	return ps.getIndex(blockID) != -1
}

func (ps *ProposalSet) getIndex(blockID BlockID) int {
	for i, b := range ps.proposals {
		if b.Block.BlockID().Equal(blockID) {
			return i
		}
	}
	return -1
}

func (ps *ProposalSet) Count() int {
	return len(ps.proposals)
}

// CertificateSet stores certificates for an epoch of consensus.
type CertificateSet struct {
	certificates []*Certificate
}

// NewCertificateSet creates a CertificateSet.
func NewCertificateSet() *CertificateSet {
	return &CertificateSet{}
}

// Get returns certificate for blockID with no epoch and height set.
func (cs *CertificateSet) Get(epoch int64, blockID BlockID, height int64) *Certificate {
	for _, c := range cs.certificates {
		if c.BlockID().Equal(blockID) && c.Height == height {
			return c
		}
	}
	return nil
}

func (cs *CertificateSet) GetVote(epoch int64, blockID BlockID, height int64, sender int) *Message {
	for _, c := range cs.certificates {
		if c.BlockID().Equal(blockID) && c.Height == height {
			return c.ReconstructMessage(sender)
		}
	}
	return nil
}

func (cs *CertificateSet) Add(c *Certificate) {
	cs.certificates = append(cs.certificates, c)
}
