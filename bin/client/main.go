package main

import (
	"flag"
	"fmt"
	"time"

	"dslab.inf.usi.ch/tendermint/net"
	"dslab.inf.usi.ch/tendermint/net/proxy"
)

var debug bool
var sd bool
var zone string
var randomSeed int64

var eid int64
var pid int
var n int
var k int

var model string

var topology string

var log net.Log
var client *proxy.Client

func init() {
	// Agent setup
	flag.BoolVar(&debug, "debug", false, "Enables debug.")
	flag.BoolVar(&sd, "sd", false, "Enables separate dissemination")
	flag.IntVar(&pid, "i", -1, "Process ID.")
	flag.IntVar(&n, "n", 0, "Number of processes.")
	flag.StringVar(&model, "mod", "sync", "Network model.")

	// Client setup
	flag.Float64Var(&rate, "rate", 1, "Rate of proposals (values/sec).")
	flag.IntVar(&size, "s", 1024, "Size of proposed values.")
	flag.IntVar(&duration, "d", 15, "Duration of the experiment in seconds.")
	flag.IntVar(&maxDuration, "dmax", 45, "Maximum duration of the experiment in seconds.")
	flag.IntVar(&perfInterval, "p", 5, "Performance stats interval in seconds.")

	// Experiment setup
	flag.Int64Var(&eid, "e", 0, "Experiment ID.")
	flag.Int64Var(&randomSeed, "seed", 0, "Random seed for the experiment. When unset, the experiment ID is used.")
	flag.StringVar(&rendezvousAddr, "r", "", "Rendevouz full addresses in Multiaddr format.")
	flag.StringVar(&topology, "topology", "", "Topology of the agents in the experiment.")
	flag.StringVar(&zone, "zone", "LAN", "Zone that hosts the agent.")
}

func main() {
	panic(fmt.Errorf("Client deprecated, do not use it!"))

	flag.Parse()

	log = net.StartLog(fmt.Sprint("c", pid))

	if len(topology) != 0 {
		if k > 0 {
			topology = "gossip"
		} else {
			topology = "star"
		}
	} else {
		topology = "full"
	}
	log.Println("Topology:", topology)

	log.Println("Network model:", model)

	SetupHostClient()
	log.Println("host:", host.AddrInfo())

	timestamp := time.Now()
	peers := FindPeers()
	count := peers.WaitN(n)
	duration := time.Now().Sub(timestamp)
	log.Println("Found", count, "peers in", duration)

	// Finish the experiment if not enough peers were found
	if count < n {
		panic(fmt.Errorf("expected", n, "peers, found", count))
	}

	proxyNamespace := fmt.Sprint(eid, "/", zone)

	var err error
	timestamp = time.Now()
	cpeers := discovery.FindPeers(proxyNamespace, 100)
	log.Println("Searching proxy in namespace", proxyNamespace)
	for cpeer := range cpeers.Queue {
		client, err = proxy.NewClient(host, cpeer)
		if err != nil {
			log.Println("failed to connect to proxy",
				cpeer, err)
			continue
		}
		duration = time.Now().Sub(timestamp)
		log.Println("connected to proxy", cpeer,
			"after", duration)
		break
	}
	// Panics if no connection with a proxy was established
	if client == nil {
		panic(fmt.Errorf("failed to connect to the proxy after %v",
			time.Now().Sub(timestamp)))
	}

	// Wait for the servers setup
	time.Sleep(5 * time.Second)

	clientLoop()
}
