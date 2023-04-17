package main

import (
	"bufio"
	"fmt"
	gnet "net"
	"os"
	"strings"
	"time"

	"dslab.inf.usi.ch/tendermint/net/libp2p"
)

var listenAddr string
var publicAddr string
var rendezvousAddr string

var host *libp2p.Host
var discovery *libp2p.Discovery

func SetupHost() {
	var err error
	cfg := new(libp2p.Config)
	if len(listenAddr) > 0 {
		cfg.ListenAddr = listenAddr
	} else {
		cfg.ListenAddr = DefaultListenAddr()
	}
	cfg.PublicAddr = publicAddr
	cfg.Identity = libp2p.DeterministicEDSAKey(int64(pid * 100))
	host, err = libp2p.NewHostWithConfig(pid, cfg)
	if err != nil {
		panic(fmt.Errorf("SetupHost: %s", err))
	}
}

func FindPeers() *libp2p.PeerList {
	var err error
	if len(rendezvousAddr) == 0 {
		rendezvousAddr = DefaultRendezvousAddr()
	}
	times := 0
	discovery, err = host.NewDiscoveryClient(rendezvousAddr)
	for err != nil && times < 10 {
		discovery, err = host.NewDiscoveryClient(rendezvousAddr)
		time.Sleep(time.Millisecond * 500)
		times++
	}
	if err != nil {
		panic(fmt.Errorf("Discovery: %s", err))
	}
	namespace := fmt.Sprint(eid)
	if len(host.AddrInfo().Addrs) > 0 {
		err = discovery.Advertise(namespace)
		if err != nil {
			panic(fmt.Errorf("Advertise(%s): %s", namespace, err))
		}
	}
	return discovery.FindPeers(namespace, n)
}

// Automatically detects the main address of cluster nodes

var HostnamePrefix = "node"

func DefaultListenAddr() string {
	hostname, err := os.Hostname()
	if err == nil && strings.HasPrefix(hostname, HostnamePrefix) {
		addrs, err := gnet.LookupHost(hostname)
		if err == nil {
			return fmt.Sprintf("/ip4/%s/tcp/%d", addrs[0], 0)
		}
	}
	return ""
}

// Retrieve the rendezvous address from the server log file, if possible

var RendezvousLogFile = "./rendezvous.log"

func DefaultRendezvousAddr() string {
	file, err := os.Open(RendezvousLogFile)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Rendezvous address: ") {
				s := strings.Split(line, " ")
				if len(s) > 2 {
					return s[2]
				}
			}
		}
	}
	return fmt.Sprint("Unable to find log file '", RendezvousLogFile, "'")
}
