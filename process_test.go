package tendermint

/*
import (
	"testing"

	"dslab.inf.usi.ch/tendermint/consensus"
)




func testProcessSet(n int) []*Process {
	p := make([]*Process, n)
	for id := 0; id < n; id++ {
		p[id] = NewProcess(id, n, nil, nil, nil)
		if p[id].ID() != id {
			panic(fmt.Sprint("Process", id, "returned ID", p[id].ID()))
		}
	}
	return p
}

func TestProcessProposerRound0(t *testing.T) {
	n := 4           // Number of processes
	hmax := int64(n) // Heights to test, minimum n
	p := testProcessSet(n)
	setAsProposer := make([]int, n)
	for h := int64(0); h < hmax; h++ {
		// Proposer is deterministic
		pr := p[0].Proposer(h, 0)
		for id := 1; id < n; id++ {
			if p := p[id].Proposer(h, 0); p != pr {
				t.Error("Proposer of height", h, "expected", pr,
					"got", p, "at process", id)
			}
		}
		setAsProposer[pr] += 1
	}

	// All processes should have their turn
	for id := 0; id < n; id++ {
		if setAsProposer[id] == 0 {
			t.Error("Process", id, "was not the round0 proposer in", hmax, "heights")
		}
	}
}

func TestProcessAcceptsIncorporatedGossip(t *testing.T) {
	g := new(gossip.Gossip)
	// p.gossip is of 'net.Gossip' interface
	p := &Process{
		gossip: g,
	}
	if p == nil {
		t.Error("Nil process")
	}
}

func TestProcessAcceptsIncorporatedProxy(t *testing.T) {
	pry := new(proxy.Proxy)
	// p.proxy is of 'net.Proxy' interface
	p := &Process{
		proxy: pry,
	}
	if p == nil {
		t.Error("Nil process")
	}
}
*/
