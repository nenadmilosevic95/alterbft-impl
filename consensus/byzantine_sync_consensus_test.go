package consensus

/*
import (
	"math/rand"
	"testing"
	"time"
)

func TestGetRandomByzantines(t *testing.T) {
	n := 25
	f := 12
	for i := 0; i < 1000; i++ {
		byzantines := getRandomByzantines(f, n, time.Now().UnixNano())
		if len(byzantines) != f {
			t.Errorf("GetRandomByzantines returns %v byzantines expected %v\n", len(byzantines), f)
		}
	}
}

func TestCheckByzantineExistInZone(t *testing.T) {
	n := 25
	byzantines := make(map[int]bool, NumberOfZones)
	if checkByzantineExistInZone(Virginia, byzantines, n) == true {
		t.Error("checkByzantineExistInZone returns true expected false\n")
	}
	byzantines[Virginia] = true
	if checkByzantineExistInZone(Virginia, byzantines, n) == false {
		t.Error("checkByzantineExistInZone returns false expected true\n")
	}

	if checkByzantineExistInZone(SaoPaolo, byzantines, n) == true {
		t.Error("checkByzantineExistInZone returns true expected false\n")
	}
	byzantines[SaoPaolo+NumberOfZones] = true
	if checkByzantineExistInZone(SaoPaolo, byzantines, n) == false {
		t.Error("checkByzantineExistInZone returns false expected true\n")
	}

	if checkByzantineExistInZone(Stockholm, byzantines, n) == true {
		t.Error("checkByzantineExistInZone returns true expected false\n")
	}
	byzantines[Stockholm+2*NumberOfZones] = true
	if checkByzantineExistInZone(Stockholm, byzantines, n) == false {
		t.Error("checkByzantineExistInZone returns false expected true\n")
	}

	if checkByzantineExistInZone(Singapore, byzantines, n) == true {
		t.Error("checkByzantineExistInZone returns true expected false\n")
	}
	byzantines[Singapore+3*NumberOfZones] = true
	if checkByzantineExistInZone(Singapore, byzantines, n) == false {
		t.Error("checkByzantineExistInZone returns false expected true\n")
	}

	if checkByzantineExistInZone(Syndey, byzantines, n) == true {
		t.Error("checkByzantineExistInZone returns true expected false\n")
	}
	byzantines[Syndey+4*NumberOfZones] = true
	if checkByzantineExistInZone(Syndey, byzantines, n) == false {
		t.Error("checkByzantineExistInZone returns false expected true\n")
	}
}

func TestCheckByzantines(t *testing.T) {
	n := 25
	byzantines := make(map[int]bool, NumberOfZones)
	if checkByzantines(byzantines, n) == true {
		t.Error("checkByzantine returns true expected false\n")
	}
	byzantines[Virginia+4*NumberOfZones] = true
	if checkByzantines(byzantines, n) == true {
		t.Error("checkByzantine returns true expected false\n")
	}
	byzantines[SaoPaolo+2*NumberOfZones] = true
	if checkByzantines(byzantines, n) == true {
		t.Error("checkByzantine returns true expected false\n")
	}
	byzantines[Stockholm+1*NumberOfZones] = true
	if checkByzantines(byzantines, n) == true {
		t.Error("checkByzantine returns true expected false\n")
	}
	byzantines[Singapore+NumberOfZones] = true
	if checkByzantines(byzantines, n) == true {
		t.Error("checkByzantine returns true expected false\n")
	}
	byzantines[Syndey+3*NumberOfZones] = true
	if checkByzantines(byzantines, n) == false {
		t.Error("checkByzantine returns false expected true\n")
	}
}

func TestGenerateByzantines(t *testing.T) {
	n := 25
	f := 12
	for i := 0; i < 10000; i++ {
		byzantines := GenerateByzantines(f, n, time.Now().UnixNano())
		if checkByzantines(byzantines, n) == false || len(byzantines) != f {
			t.Errorf("Byzantine set doesn't cover all zones %v\n", byzantines)
		}
	}
}

func TestGetHonestProcessInZone(t *testing.T) {
	n := 25
	f := 12
	for i := 0; i < 10000; i++ {
		byzantines := GenerateByzantines(f, n, time.Now().UnixNano())
		c := NewByzantineSyncConsensus(0, nil, nil, 0, "silence")
		id := c.getHonestProcessInZone(Virginia, byzantines, n)
		if id%NumberOfZones != Virginia {
			t.Errorf("GetHonest returns %v in wrong zone.\n", id)
		}
		if byzantines[id] == true {
			t.Errorf("GetHonest returns %v that is byzantine.\n", id)
		}
		id = c.getHonestProcessInZone(SaoPaolo, byzantines, n)
		if id%NumberOfZones != SaoPaolo {
			t.Errorf("GetHonest returns %v in wrong zone.\n", id)
		}
		if byzantines[id] == true {
			t.Errorf("GetHonest returns %v that is byzantine.\n", id)
		}
		id = c.getHonestProcessInZone(Stockholm, byzantines, n)
		if id%NumberOfZones != Stockholm {
			t.Errorf("GetHonest returns %v in wrong zone.\n", id)
		}
		if byzantines[id] == true {
			t.Errorf("GetHonest returns %v that is byzantine.\n", id)
		}
		id = c.getHonestProcessInZone(Singapore, byzantines, n)
		if id%NumberOfZones != Singapore {
			t.Errorf("GetHonest returns %v in wrong zone.\n", id)
		}
		if byzantines[id] == true {
			t.Errorf("GetHonest returns %v that is byzantine.\n", id)
		}
		id = c.getHonestProcessInZone(Syndey, byzantines, n)
		if id%NumberOfZones != Syndey {
			t.Errorf("GetHonest returns %v in wrong zone.\n", id)
		}
		if byzantines[id] == true {
			t.Errorf("GetHonest returns %v that is byzantine.\n", id)
		}
	}
}

func TestClosestByzantineProcess(t *testing.T) {
	n := 25
	f := 12
	for i := 0; i < 10000; i++ {
		byzantines := GenerateByzantines(f, n, time.Now().UnixNano())
		c := NewByzantineSyncConsensus(0, nil, nil, 0, "silence")
		id := c.getHonestProcessInZone(Virginia, byzantines, n)
		b := c.closestByzantineProcess(id, byzantines, n)
		if b%NumberOfZones != Virginia {
			t.Errorf("ClosestByzantineProcess returns %v in wrong zone.\n", id)
		}
		if byzantines[b] == false {
			t.Errorf("ClosestByzantineProcess returns %v that is not byzantine.\n", id)
		}
		id = c.getHonestProcessInZone(SaoPaolo, byzantines, n)
		b = c.closestByzantineProcess(id, byzantines, n)
		if b%NumberOfZones != SaoPaolo {
			t.Errorf("ClosestByzantineProcess returns %v in wrong zone.\n", id)
		}
		if byzantines[b] == false {
			t.Errorf("ClosestByzantineProcess returns %v that is not byzantine.\n", id)
		}
		id = c.getHonestProcessInZone(Stockholm, byzantines, n)
		b = c.closestByzantineProcess(id, byzantines, n)
		if b%NumberOfZones != Stockholm {
			t.Errorf("ClosestByzantineProcess returns %v in wrong zone.\n", id)
		}
		if byzantines[b] == false {
			t.Errorf("ClosestByzantineProcess returns %v that is not byzantine.\n", id)
		}
		id = c.getHonestProcessInZone(Singapore, byzantines, n)
		b = c.closestByzantineProcess(id, byzantines, n)
		if b%NumberOfZones != Singapore {
			t.Errorf("ClosestByzantineProcess returns %v in wrong zone.\n", id)
		}
		if byzantines[b] == false {
			t.Errorf("ClosestByzantineProcess returns %v that is not byzantine.\n", id)
		}
		id = c.getHonestProcessInZone(Syndey, byzantines, n)
		b = c.closestByzantineProcess(id, byzantines, n)
		if b%NumberOfZones != Syndey {
			t.Errorf("ClosestByzantineProcess returns %v in wrong zone.\n", id)
		}
		if byzantines[b] == false {
			t.Errorf("ClosestByzantineProcess returns %v that is not byzantine.\n", id)
		}
	}
}

func TestFurthestProcess(t *testing.T) {
	n := 25
	f := 12
	byzantines := GenerateByzantines(f, n, time.Now().UnixNano())
	c := NewByzantineSyncConsensus(0, nil, nil, 0, "silence")
	p := c.getHonestProcessInZone(Virginia, byzantines, n)
	id := c.furthestProcess(p, byzantines, n)
	if id%NumberOfZones != Singapore {
		t.Errorf("FurthestProcess returns %v in wrong zone.\n", id)
	}
	if byzantines[id] == true {
		t.Errorf("FurthessProcess returns %v that is byzantine.\n", id)
	}
	p = c.getHonestProcessInZone(SaoPaolo, byzantines, n)
	id = c.furthestProcess(p, byzantines, n)
	if id%NumberOfZones != Singapore {
		t.Errorf("FurthestProcess returns %v in wrong zone.\n", id)
	}
	if byzantines[id] == true {
		t.Errorf("FurthessProcess returns %v that is byzantine.\n", id)
	}
	p = c.getHonestProcessInZone(Stockholm, byzantines, n)
	id = c.furthestProcess(p, byzantines, n)
	if id%NumberOfZones != Syndey {
		t.Errorf("FurthestProcess returns %v in wrong zone.\n", id)
	}
	if byzantines[id] == true {
		t.Errorf("FurthessProcess returns %v that is byzantine.\n", id)
	}
	p = c.getHonestProcessInZone(Singapore, byzantines, n)
	id = c.furthestProcess(p, byzantines, n)
	if id%NumberOfZones != SaoPaolo {
		t.Errorf("FurthestProcess returns %v in wrong zone.\n", id)
	}
	if byzantines[id] == true {
		t.Errorf("FurthessProcess returns %v that is byzantine.\n", id)
	}
	p = c.getHonestProcessInZone(Syndey, byzantines, n)
	id = c.furthestProcess(p, byzantines, n)
	if id%NumberOfZones != SaoPaolo {
		t.Errorf("FurthestProcess returns %v in wrong zone.\n", id)
	}
	if byzantines[id] == true {
		t.Errorf("FurthessProcess returns %v that is byzantine.\n", id)
	}
}

func TestTwoMostDistantProcesses(t *testing.T) {
	n := 25
	f := 12
	byzantines := GenerateByzantines(f, n, time.Now().UnixNano())
	c := NewByzantineSyncConsensus(0, nil, nil, 0, "silence")
	p, q := c.twoMostDistantProcesses(byzantines, n)
	if p%NumberOfZones != SaoPaolo {
		t.Errorf("MostDistant returns p %v in wrong zone.\n", p)
	}
	if byzantines[p] == true {
		t.Errorf("MostDistant returns p %v that is byzantine.\n", p)
	}
	if q%NumberOfZones != Singapore {
		t.Errorf("MostDistant returns q %v in wrong zone.\n", q)
	}
	if byzantines[q] == true {
		t.Errorf("MostDistant returns q %v that is byzantine.\n", q)
	}
}

func TestGetPandQ(t *testing.T) {
	n := 25
	f := 12
	byzantines := GenerateByzantines(f, n, time.Now().UnixNano())
	c := NewByzantineSyncConsensus(0, nil, nil, 0, "silence")
	// Honest leader
	leader := c.getHonestProcessInZone(Virginia, byzantines, n)
	p, q := c.getPandQ(leader, byzantines, n)
	if p != leader {
		t.Errorf("GetPandQ returns p %v expected %v\n", p, leader)
	}
	qq := c.furthestProcess(leader, byzantines, n)
	if q != qq {
		t.Errorf("GetPandQ returns q %v expected %v\n", q, qq)
	}
	leader = c.getHonestProcessInZone(SaoPaolo, byzantines, n)
	p, q = c.getPandQ(leader, byzantines, n)
	if p != leader {
		t.Errorf("GetPandQ returns p %v expected %v\n", p, leader)
	}
	qq = c.furthestProcess(leader, byzantines, n)
	if q != qq {
		t.Errorf("GetPandQ returns q %v expected %v\n", q, qq)
	}
	leader = c.getHonestProcessInZone(Stockholm, byzantines, n)
	p, q = c.getPandQ(leader, byzantines, n)
	if p != leader {
		t.Errorf("GetPandQ returns p %v expected %v\n", p, leader)
	}
	qq = c.furthestProcess(leader, byzantines, n)
	if q != qq {
		t.Errorf("GetPandQ returns q %v expected %v\n", q, qq)
	}
	leader = c.getHonestProcessInZone(Singapore, byzantines, n)
	p, q = c.getPandQ(leader, byzantines, n)
	if p != leader {
		t.Errorf("GetPandQ returns p %v expected %v\n", p, leader)
	}
	qq = c.furthestProcess(leader, byzantines, n)
	if q != qq {
		t.Errorf("GetPandQ returns q %v expected %v\n", q, qq)
	}
	leader = c.getHonestProcessInZone(Syndey, byzantines, n)
	p, q = c.getPandQ(leader, byzantines, n)
	if p != leader {
		t.Errorf("GetPandQ returns p %v expected %v\n", p, leader)
	}
	qq = c.furthestProcess(leader, byzantines, n)
	if q != qq {
		t.Errorf("GetPandQ returns q %v expected %v\n", q, qq)
	}
	// Byzantine leader
	leader = c.closestByzantineProcess(0, byzantines, n)
	p, q = c.getPandQ(leader, byzantines, n)
	pp := c.getHonestProcessInZone(SaoPaolo, byzantines, n)
	qq = c.getHonestProcessInZone(Singapore, byzantines, n)
	if p != pp || q != qq {
		t.Errorf("GetPandQ returns %v and %v expected %v and %v\n", p, q, pp, qq)
	}
	leader = c.closestByzantineProcess(1, byzantines, n)
	p, q = c.getPandQ(leader, byzantines, n)
	pp = c.getHonestProcessInZone(SaoPaolo, byzantines, n)
	qq = c.getHonestProcessInZone(Singapore, byzantines, n)
	if p != pp || q != qq {
		t.Errorf("GetPandQ returns %v and %v expected %v and %v\n", p, q, pp, qq)
	}
	leader = c.closestByzantineProcess(2, byzantines, n)
	p, q = c.getPandQ(leader, byzantines, n)
	pp = c.getHonestProcessInZone(SaoPaolo, byzantines, n)
	qq = c.getHonestProcessInZone(Singapore, byzantines, n)
	if p != pp || q != qq {
		t.Errorf("GetPandQ returns %v and %v expected %v and %v\n", p, q, pp, qq)
	}
	leader = c.closestByzantineProcess(3, byzantines, n)
	p, q = c.getPandQ(leader, byzantines, n)
	pp = c.getHonestProcessInZone(SaoPaolo, byzantines, n)
	qq = c.getHonestProcessInZone(Singapore, byzantines, n)
	if p != pp || q != qq {
		t.Errorf("GetPandQ returns %v and %v expected %v and %v\n", p, q, pp, qq)
	}
	leader = c.closestByzantineProcess(4, byzantines, n)
	p, q = c.getPandQ(leader, byzantines, n)
	pp = c.getHonestProcessInZone(SaoPaolo, byzantines, n)
	qq = c.getHonestProcessInZone(Singapore, byzantines, n)
	if p != pp || q != qq {
		t.Errorf("GetPandQ returns %v and %v expected %v and %v\n", p, q, pp, qq)
	}
}

func TestGetSilences(t *testing.T) {
	n := 25
	f := 12
	e := int64(2)
	byzantines := GenerateByzantines(f, n, time.Now().UnixNano())
	c := NewByzantineSyncConsensus(e, nil, nil, 0, "silence")
	msgs := c.getSilenceMessages(byzantines)
	if len(msgs) != f {
		t.Errorf("GetSilenceMessages returns %v msgs instead of %v\n", len(msgs), f)
	}
	i := 0
	for _, m := range msgs {
		if byzantines[m.Sender] == false {
			t.Errorf("GetSilenceMessages returns msg with honest sender: %v.", m)
		}
		mm := NewSilenceMessage(e, int16(m.Sender))
		assertMessageEquals(t, m, mm)
		i++
	}
}

func TestGetVotes(t *testing.T) {
	n := 25
	f := 12
	e := int64(2)
	byzantines := GenerateByzantines(f, n, time.Now().UnixNano())
	c := NewByzantineSyncConsensus(e, nil, nil, 0, "silence")
	blockID := make([]byte, BlockIDSize)
	rand.Seed(time.Now().UnixNano())
	rand.Read(blockID)
	msgs := c.getVoteMessages(byzantines, blockID)
	if len(msgs) != f {
		t.Errorf("GetVoteMessages returns %v msgs instead of %v\n", len(msgs), f)
	}
	i := 0
	for _, m := range msgs {
		if byzantines[m.Sender] == false {
			t.Errorf("GetVoteMessages returns msg with honest sender: %v.", m)
		}
		mm := NewVoteMessage(e, blockID, int16(m.Sender))
		assertMessageEquals(t, m, mm)
		i++
	}
}
*/
