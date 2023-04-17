package main

import (
	"flag"
	"fmt"
	"os"

	"dslab.inf.usi.ch/tendermint/net/libp2p"
)

var listenAddr string
var publicAddr string

var outFileName string

var eid int64

func init() {
	flag.StringVar(&listenAddr, "l", "",
		"Host listen adddress in Multiaddr format")
	flag.StringVar(&publicAddr, "lp", "",
		"Host public/external adddress in Multiaddr format")
	flag.StringVar(&outFileName, "o", "", "Output file name.")
	flag.Int64Var(&eid, "e", 0,
		"Experiment ID, used to generate the server's host ID.")
}

func main() {
	flag.Parse()

	var cfg libp2p.Config
	if len(listenAddr) > 0 {
		cfg.ListenAddr = listenAddr
	}
	if len(publicAddr) > 0 {
		cfg.PublicAddr = publicAddr
	}
	if eid > 0 {
		cfg.Identity = libp2p.DeterministicEDSAKey(eid)
	} // Else: host Identity will be randomly generated

	var outFile *os.File
	if len(outFileName) > 0 {
		var err error
		outFile, err = os.Create(outFileName)
		if err != nil {
			panic(err)
		}
	}

	host, err := libp2p.NewHostWithConfig(0, &cfg)
	if err != nil {
		panic(err)
	}

	_, err = host.NewDiscoveryServer()
	if err != nil {
		panic(err)
	}

	addrInfo := host.AddrInfo()
	for _, addr := range addrInfo.Addrs {
		addrStr := fmt.Sprintf("Rendezvous address: %s/p2p/%s",
			addr, addrInfo.ID)
		fmt.Println(addrStr)
		if outFile != nil {
			fmt.Fprintln(outFile, addrStr)
		}
	}

	select {}
}
