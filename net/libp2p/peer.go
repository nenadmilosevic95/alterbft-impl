package libp2p

import (
	"math/rand"
	"sort"
	"sync"

	"github.com/libp2p/go-libp2p-core/peer"
)

type PeerList struct {
	Namespace string
	Queue     chan peer.AddrInfo

	lock  sync.Mutex
	list  []peer.AddrInfo
	table map[peer.ID]*peer.AddrInfo
}

// NewPeerList creates a list of peers with capacitity to number peers.
// A peer list is fed by the peer discovery service, that looks for peers
// in the provided namespace. Peers found are added to the Queue channel.
func NewPeerList(namespace string, number int) *PeerList {
	return &PeerList{
		Namespace: namespace,
		Queue:     make(chan peer.AddrInfo, number),

		list:  make([]peer.AddrInfo, 0, number),
		table: make(map[peer.ID]*peer.AddrInfo, number),
	}
}

// AddPeer adds a peer to the peer list, if not already added.
func (p *PeerList) AddPeer(peer peer.AddrInfo) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.table[peer.ID] != nil {
		return false
	}
	p.list = append(p.list, peer)
	p.table[peer.ID] = &peer
	return true
}

// GetPeer returns the provided peer, if it is found, or nil.
func (p *PeerList) GetPeer(peerID peer.ID) *peer.AddrInfo {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.table[peerID]
}

// List returns the list of peers found.
func (p *PeerList) List() []peer.AddrInfo {
	p.lock.Lock()
	defer p.lock.Unlock()
	var peers []peer.AddrInfo
	for i := range p.list {
		peers = append(peers, p.list[i])
	}
	return peers
}

// SortByUID returns the list of peers found, sorted by their libp2p IDs.
func (p *PeerList) SortByPeerID() []peer.AddrInfo {
	peers := p.List()
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].ID < peers[j].ID
	})
	return peers
}

// Shuffle returns the list of peers found, in a random order.
func (p *PeerList) Shuffle() []peer.AddrInfo {
	peers := p.SortByPeerID()
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})
	return peers
}

// WaitN blocks until the provided number of peers are found.
// Assumes that new peers are added to the PeerList's Queue channel.
func (p *PeerList) WaitN(number int) int {
	var found int
	for _ = range p.Queue {
		found += 1
		if found >= number {
			break
		}
	}
	return found
}
