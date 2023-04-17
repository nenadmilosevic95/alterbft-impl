package libp2p

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
)

// AddrInfo parses a libp2p's peer address information from a string.
func AddrInfo(address string) peer.AddrInfo {
	addr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return peer.AddrInfo{}
	}
	addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return peer.AddrInfo{}
	}
	return *addrInfo
}

// Protocol parses a libp2p's protocol.ID instance from a string.
func Protocol(protocolID string) protocol.ID {
	return protocol.ID(protocolID)
}
