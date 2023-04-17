package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dslab.inf.usi.ch/tendermint/net/libp2p"
)

var listenAddr string
var rendezvousAddr string

var host *libp2p.Host
var discovery *libp2p.Discovery

func SetupHostClient() {
	var err error
	cfg := new(libp2p.Config)
	cfg.ListenAddr = "none"
	host, err = libp2p.NewHostWithConfig(pid, cfg)
	if err != nil {
		panic(fmt.Errorf("SetupHostClient: %s", err))
	}
}

func FindPeers() *libp2p.PeerList {
	var err error
	if len(rendezvousAddr) == 0 {
		rendezvousAddr = DefaultRendezvousAddr()
	}
	discovery, err = host.NewDiscoveryClient(rendezvousAddr)
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
