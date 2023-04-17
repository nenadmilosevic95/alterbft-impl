package libp2p

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var (
	DiscoveryAdvertiseTTL  = time.Minute
	DiscoveryQueryInterval = 100 * time.Millisecond
	DiscoveryQueryTimeout  = 10 * time.Second
)

type Discovery struct {
	*Host // Inherit fields and methods

	Dht *dht.IpfsDHT
}

// NewDiscovery creates a Discovery instance associated with the provide Host.
// The instantiated DHT daemon can operate either in client or server mode.
func NewDiscovery(host *Host, client bool) (*Discovery, error) {
	var discovery *Discovery
	var modeOption dht.Option
	if client {
		modeOption = dht.Mode(dht.ModeClient)
	} else {
		modeOption = dht.Mode(dht.ModeServer)
	}
	dht, err := dht.New(host.Ctx, host.Host, modeOption)
	if err == nil {
		discovery = &Discovery{host, dht}
	}
	return discovery, err
}

// Advertise the host in the provided namespace.
// This allows other peers to discovery the host in the same namespace.
func (d *Discovery) Advertise(namespace string) error {
	ttl := discovery.TTL(DiscoveryAdvertiseTTL)
	router := routing.NewRoutingDiscovery(d.Dht)
	ctx, _ := context.WithTimeout(d.Ctx, DiscoveryQueryTimeout)
	_, err := router.Advertise(ctx, namespace, ttl)

	// The kbucket.ErrLookupFailure error is very common and not critical:
	// while this error is generated, try again advertising in namespace.
	for err != nil && err == kbucket.ErrLookupFailure {
		time.Sleep(DiscoveryQueryInterval)
		_, err = router.Advertise(ctx, namespace, ttl)
	}
	return err
}

// FindPeers looks to up to n peers that announced in the provided namespace.
// The method does not block: peers found are added to the returned peer list.
func (d *Discovery) FindPeers(namespace string, n int) *PeerList {
	peers := NewPeerList(namespace, n)
	go d.findPeersAsyncRoutine(peers, n)
	return peers
}

// Blocking method that looks to up to n peers and adds them to a peer list.
// The method returns when up to n distinct peers are found or, irrespective
// of the number of peers found, after DiscoveryQueryTimeout time units.
func (d *Discovery) findPeersAsyncRoutine(peers *PeerList, n int) {
	var found int
	limit := discovery.Limit(n)
	router := routing.NewRoutingDiscovery(d.Dht)
	deadline := time.Now().Add(DiscoveryQueryTimeout)
	ctx, _ := context.WithTimeout(d.Ctx, DiscoveryQueryTimeout)
	for found < n && time.Now().Before(deadline) {
		list, err := router.FindPeers(ctx, peers.Namespace, limit)
		if err != nil {
			break
		}

		for peer := range list {
			// Duplicated peers can be found
			if peers.AddPeer(peer) {
				found += 1
				peers.Queue <- peer
			}
		}

		if found < n {
			time.Sleep(DiscoveryQueryInterval)
		}
	}
	close(peers.Queue)
}
