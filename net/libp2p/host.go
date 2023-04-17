package libp2p

import (
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
)

// Host encapsulate a lip2p's Host entity with a given Paxos ID.
type Host struct {
	ID int // Paxos ID

	Cfg  Config
	Ctx  context.Context
	Host host.Host
}

// Create a Host with default parameters.
func NewHost(id int) (*Host, error) {
	return NewHostWithConfig(id, DefaultConfig())
}

// Create a Host with the provided config parameters
func NewHostWithConfig(id int, cfg *Config) (*Host, error) {
	if cfg.Context == nil {
		cfg.Context = context.Background()
	}
	// Context is not part of the constructor anymore
	libp2pHost, err := libp2p.New(cfg.LoadLibp2pOptions()...)
	if err != nil {
		return nil, err
	}
	host := &Host{
		ID:   id,
		Cfg:  *cfg,
		Ctx:  cfg.Context,
		Host: libp2pHost,
	}
	return host, nil
}

// AddrInfo returns the host's libp2p's ID and listen addresses.
func (h *Host) AddrInfo() peer.AddrInfo {
	return *host.InfoFromHost(h.Host)

}

// Connect establishes a connection with the remote host.
func (h *Host) Connect(remote peer.AddrInfo) error {
	return h.Host.Connect(h.Ctx, remote)
}

// NewDiscoveryClient creates a Discovery client instance for peer discovery.
// Also connects to the provided rendezvous address, i.e., Discovery server.
func (h *Host) NewDiscoveryClient(rendezvousAddr string) (*Discovery, error) {
	discovery, err := NewDiscovery(h, true)
	if err != nil {
		return nil, err
	}
	err = h.Connect(AddrInfo(rendezvousAddr))
	return discovery, err
}

// NewDiscoveryServer creates a Discovery server instance for peer discovery.
func (h *Host) NewDiscoveryServer() (*Discovery, error) {
	return NewDiscovery(h, false)
}

// NewStream creates a stream with the remote host using the provided protocol.
// A connection with the remote host is established when necessary.
func (h *Host) NewStream(remote peer.AddrInfo, protocol protocol.ID) (network.Stream, error) {
	// Ensure that the host knows the remote address
	h.Host.Peerstore().AddAddrs(remote.ID, remote.Addrs,
		peerstore.AddressTTL)
	return h.Host.NewStream(h.Ctx, remote.ID, protocol)
}

// UID returns the libp2p's host unique identifier.
func (h *Host) UID() peer.ID {
	return h.Host.ID()
}
