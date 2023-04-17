package main

import (
	"time"

	"dslab.inf.usi.ch/tendermint/net/gossip"
	"github.com/libp2p/go-libp2p-core/peer"
)

var gtransport *gossip.Gossip
var gdonechan chan *Gdone

var gossipSetupTimeout = 10 * time.Second

func SetupGossip() {
	// Ensure proper maximum message size
	if size > gossip.MaxPayloadSize {
		gossip.MaxPayloadSize = size
	}

	gtransport = gossip.NewGossipTransport(host, log, msgLossRate)
	gdonechan = make(chan *Gdone)
	go GossipMonitor()
}

func SetupStar() {
	// Ensure proper maximum message size
	if size > gossip.MaxPayloadSize {
		gossip.MaxPayloadSize = size
	}

	gtransport = gossip.NewUnicastTransport(host, log, msgLossRate)
	gdonechan = make(chan *Gdone)
	go GossipMonitor()
}

type Gdone struct {
	Done chan int
	K    int
}

func GossipConnect(peers []peer.AddrInfo, target int) {
	var selected int
	for _, peer := range peers {
		if peer.ID == host.AddrInfo().ID {
			continue
		}
		if len(peer.Addrs) == 0 {
			log.Println("GossipConnect ignoring", peer)
			continue
		}
		selected += 1
		gtransport.Network.Connect(peer)
		if selected >= target {
			break
		}
	}
}

func GossipWait(numberOfPeers int) *Gdone {
	gdone := &Gdone{
		Done: make(chan int),
		K:    numberOfPeers,
	}
	gdonechan <- gdone
	return gdone
}

func GossipMonitor() {
	var gdone *Gdone
	peers := gtransport.Peers
	peers.Lock()
	peers.Notifier()
	peers.Unlock()
	active := make(map[peer.ID]bool)
	errors := make(map[peer.ID]bool)
	var timeout bool
	var timer <-chan time.Time
	for {
		select {
		case peer := <-peers.Events:
			rid := peer.Addr.ID
			if peer.Errors() || errors[rid] {
				errors[rid] = true
				logPeerState(&peer)
			} else if debug {
				logPeerState(&peer)
			}
			if peer.FulllyConnected() && !active[rid] {
				active[rid] = true
				delete(errors, rid)
				logPeerActive(&peer)
			}

		case gdone = <-gdonechan:
			timer = time.After(gossipSetupTimeout)
		case <-timer:
			timeout = true
		}

		if gdone != nil &&
			(len(active) >= gdone.K || timeout) {
			gdone.Done <- len(active)
			gdone = nil
		}
	}
}

func logPeerActive(peer *gossip.Peer) {
	log.Println("active peer", peer.ID, peer.Chosen, peer.Addr)
}

func logPeerState(peer *gossip.Peer) {
	log.Println(peer.ID, peer.Chosen, peer.Addr, peer.Comment,
		"Conn:", peer.ConnS, peer.ConnE,
		"Recv:", peer.RecvStreamS, peer.RecvStreamE,
		"Send:", peer.SendStreamS, peer.SendStreamE)
}
