package gossip

import (
	"bufio"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Peer struct {
	Addr peer.AddrInfo
	ID   int

	Active bool
	Chosen bool

	Conn  network.Conn
	ConnE error
	ConnS int

	RecvStream  network.Stream
	RecvStreamE error
	RecvStreamS int

	SendStream  network.Stream
	SendStreamE error
	SendStreamS int

	Comment string
}

func (p *Peer) BufferedReader() *bufio.Reader {
	return bufio.NewReader(p.RecvStream)
}

func (p *Peer) BufferedWriter() *bufio.Writer {
	return bufio.NewWriter(p.SendStream)
}

func (p *Peer) Errors() bool {
	return p.ConnE != nil || p.SendStreamE != nil || p.RecvStreamE != nil
}

func (p *Peer) FulllyConnected() bool {
	return p.ConnS == 2 && p.RecvStreamS == 2 && p.SendStreamS == 2
}
