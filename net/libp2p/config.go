package libp2p

import (
	"context"
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
)

type Config struct {
	// Mandatory parameter
	Context context.Context

	// Optional parameters
	ListenAddr string
	PublicAddr string
	Identity   crypto.PrivKey
}

func DefaultConfig() *Config {
	return &Config{
		Context: context.Background(),

		ListenAddr: ListenAddrDefault(),
	}
}

func (c *Config) LoadLibp2pOptions() []libp2p.Option {
	var opts []libp2p.Option
	opts = append(opts, c.loadListenAddrOption())
	if len(c.PublicAddr) > 0 {
		opts = append(opts, c.loadAddressFactoryOption())
	}
	if c.Identity != nil {
		opts = append(opts, libp2p.Identity(c.Identity))
	}
	return opts
}

func (c *Config) loadAddressFactoryOption() libp2p.Option {
	addressFactory := func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
		maddress, err := multiaddr.NewMultiaddr(c.PublicAddr)
		if err == nil {
			addrs = append(addrs, maddress)
		}
		return addrs
	}
	return libp2p.AddrsFactory(addressFactory)
}

func (c *Config) loadListenAddrOption() libp2p.Option {
	switch c.ListenAddr {
	case "": // Unset, so use the default
		return libp2p.ListenAddrStrings(ListenAddrDefault())
	case "all":
		return libp2p.DefaultListenAddrs
	case "none":
		return libp2p.NoListenAddrs
	default: // User-defined value
		return libp2p.ListenAddrStrings(c.ListenAddr)
	}
}

// ListenAddr returned consists of all addresses, the libp2p's default.
func ListenAddrAll() string {
	return "all"
}

// ListenAddr returned is the default: localhost with a random TCP port.
func ListenAddrDefault() string {
	return "/ip4/127.0.0.1/tcp/0"
}

// ListenAddr returned from a DNS query using the provided hostname.
// Returns the first retrieved IP address with a random TCP port.
// If the DNS lookup fails, returns the default ListenAddr.
func ListenAddrDNS(hostname string) string {
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return ""
	}
	return fmt.Sprint("/ip4/", addrs[0], "/tcp/0")
}

// ListenAddr returned configure the host to have no listen address.
func ListenAddrNone() string {
	return "none"
}
