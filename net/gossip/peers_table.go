package gossip

import (
	"sync"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

type PeersTable struct {
	lock  sync.Mutex
	table map[peer.ID]*Peer

	Active chan Peer
	Events chan Peer
}

func NewPeersTable() *PeersTable {
	return &PeersTable{
		table:  make(map[peer.ID]*Peer),
		Active: make(chan Peer, DefaultQueueSize),
	}
}

func (p *PeersTable) Activate(peer *Peer) {
	peer.Active = true
	p.Active <- *peer
}

func (p *PeersTable) Deactivate(peer *Peer) {
	peer.Active = false
	p.Active <- *peer
}

func (p *PeersTable) GetByAddr(addr peer.AddrInfo) *Peer {
	peer := p.table[addr.ID]
	if peer == nil {
		peer = &Peer{
			Addr: addr,
		}
		p.table[addr.ID] = peer
	}
	return peer
}

func (p *PeersTable) GetByConn(conn network.Conn) *Peer {
	return p.GetByAddr(addrInfoFromConn(conn))
}

func (p *PeersTable) GetByStream(stream network.Stream) *Peer {
	return p.GetByConn(stream.Conn())
}

func (p *PeersTable) Lock() {
	p.lock.Lock()
}

func (p *PeersTable) Notifier() chan Peer {
	p.Events = make(chan Peer, DefaultQueueSize)
	return p.Events
}

func (p *PeersTable) Notify(peer *Peer) {
	if p.Events != nil {
		p.Events <- *peer
	}
}

func (p *PeersTable) Unlock() {
	p.lock.Unlock()
}
