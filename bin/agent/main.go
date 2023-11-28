package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"dslab.inf.usi.ch/tendermint"
	"dslab.inf.usi.ch/tendermint/net"
	"dslab.inf.usi.ch/tendermint/net/gossip"
	"dslab.inf.usi.ch/tendermint/net/libp2p"
	"dslab.inf.usi.ch/tendermint/net/proxy"
	"dslab.inf.usi.ch/tendermint/workload"
)

var clientMode bool
var debug bool
var sd bool
var zone string
var advertiseProxy bool
var randomSeed int64

var eid int64
var pid int
var n int
var k int

var capacity int
var ibcap int
var obcap int
var fvcap int

// MsgLossRate introduces message loss rate
var msgLossRate float64

var topology string
var semanticFiltering bool

var model string
var smallDelta int
var bigDelta int
var fastOpt bool
var coolTime int

var numByzantines int
var byzTime int
var byzAttack string

var maxEpoch int64

var chunksNumber int

var log net.Log
var cproxy *proxy.Proxy

var process *tendermint.Process

func init() {
	flag.BoolVar(&clientMode, "client", false, "Operate in client-mode.")
	flag.BoolVar(&debug, "debug", false, "Enables debug.")
	flag.BoolVar(&sd, "sd", false, "Enables separate dissemination")
	flag.Int64Var(&eid, "e", 0, "Experiment ID.")
	flag.IntVar(&pid, "i", -1, "Process ID.")
	flag.IntVar(&n, "n", 0, "Number of processes.")
	flag.IntVar(&k, "k", 0, "Target number of neighbors.")
	flag.StringVar(&model, "mod", "alter", "Network model.")
	flag.BoolVar(&fastOpt, "fast", false, "Enable FastAlter optimization. ")

	flag.IntVar(&numByzantines, "byz", 0, "Number of byzantines.")
	flag.IntVar(&byzTime, "byzTime", 0, "Time a byzantine leader should wait.")
	flag.StringVar(&byzAttack, "attack", "silence", "Byzantine attack.")

	// Agent setup
	flag.IntVar(&capacity, "cap", 1024, "Capacity of all storages.")
	flag.IntVar(&fvcap, "fvcap", 1024, "Full value storage capacity.")
	flag.IntVar(&smallDelta, "s-delta", 150, "Sync delta in milliseconds.")
	flag.IntVar(&bigDelta, "b-delta", 1000, "Sync delta in milliseconds.")
	flag.IntVar(&coolTime, "cool", 10, "Cool down time in seconds.")
	// Host and discovery setup
	flag.StringVar(&listenAddr, "l", "",
		"Host listen adddress in Multiaddr format")
	flag.StringVar(&publicAddr, "lp", "",
		"Host public/external adddress in Multiaddr format")
	flag.StringVar(&rendezvousAddr, "r", "",
		"Rendevouz full addresses in Multiaddr format.")

	//experiment setup
	flag.StringVar(&zone, "zone", "LAN", "Zone that hosts the agent.")
	flag.BoolVar(&advertiseProxy, "proxy", true, "Advertise as a proxy for the agent's zone.")
	flag.Int64Var(&randomSeed, "seed", 0, "Random seed for the experiment. When unset, the experiment ID is used.")
	flag.StringVar(&topology, "topology", "", "Topology of the agents in the experiment.")
	flag.Int64Var(&maxEpoch, "maxEpoch", 100, "Maximum number of epochs to run in the experiment.")
	flag.IntVar(&chunksNumber, "cNum", 64, "Number of chunks.")

	// Gossip filtering parameters
	flag.IntVar(&gossip.LRUCacheSize, "gcache", 262144, "Gossip LRU cache size.")
	flag.BoolVar(&semanticFiltering, "sfilter", false, "Enables semantic filtering via gossip validator.")

	// Gossip queues setup
	flag.IntVar(&gossip.SendQueuesBatchMax, "sb", 0,
		"Send queue's maximum batch size for aggregation.")
	flag.IntVar(&gossip.SendQueuesBatchMin, "sbmin", 0,
		"Send queue's minimum batch size for aggregation.")
	flag.IntVar(&gossip.DefaultQueueSize, "qsize", 1024, "Default size for all queues.")
	flag.IntVar(&gossip.BroadcastQueueSize, "bqsize", 32, "Size of broadcast queue.")
	flag.IntVar(&gossip.DeliveryQueueSize, "dqsize", 8192, "Size of delivery queue.")
	flag.IntVar(&gossip.SendQueuesSize, "sqsize", 65536, "Size of send queues.")
	flag.BoolVar(&gossip.SendQueuesDrop, "sqdrop", false, "Set to true for send queues to drop messages when full.")
	flag.IntVar(&gossip.RecvQueueSize, "rqsize", 524288, "Size of receive queue.")
	flag.BoolVar(&gossip.RecvQueueDrop, "rqdrop", false, "Set to true for receive queue to drop messages when full.")
	flag.Float64Var(&msgLossRate, "msgloss", 0.0, "Message loss rate.")
}

func main() {
	flag.Parse()

	log = net.StartLog(fmt.Sprint("p", pid))
	//go forceExit(maxDuration)

	if len(topology) == 0 {
		topology = "full"
	}
	log.Printf("System size: %v\n", n)

	//log.Println("Topology:", topology)

	log.Printf("Model: %v\n", model)
	log.Printf("FastAlterOptimization enabled: %v\n", fastOpt)

	log.Printf("Small delta: %v\n", smallDelta)
	log.Printf("Big delta: %v\n", bigDelta)

	log.Printf("Chunk number: %v\n", chunksNumber)

	log.Printf("Number of byzantines: %v\n", numByzantines)

	log.Printf("Byz time: %v\n", byzTime)

	log.Printf("Block size(B): %v\n", size)

	if len(cpuprofile) > 0 || len(memprofile) > 0 {
		go profilerLoop()
	}

	SetupHost()
	gossip.StatsInterval = 4 * time.Second
	if topology == "gossip" {
		SetupGossip()
	} else {
		SetupStar()
	}
	cproxy = proxy.NewProxy(host, log, debug)
	log.Println("host:", host.AddrInfo())

	log.Println("[transport] broadcast queue size:",
		gossip.BroadcastQueueSize)
	log.Println("[transport] delivery queue size:",
		gossip.DeliveryQueueSize)
	log.Println("[transport] receive queue size:", gossip.RecvQueueSize,
		"drop messages:", gossip.RecvQueueDrop)
	log.Println("[transport] send queues size:", gossip.SendQueuesSize,
		"drop messages:", gossip.SendQueuesDrop)

	libp2p.DiscoveryQueryTimeout = 20 * time.Second

	timestamp := time.Now()
	peers := FindPeers()
	count := peers.WaitN(n)
	fduration := time.Now().Sub(timestamp)
	log.Println("Found", count, "peers in", fduration)

	// Finish the experiment if not enough peers were found
	if count < n {
		panic(fmt.Sprint("expected", n, "peers, found", count))
	}

	timestamp = time.Now()
	setupRandomGenerator()

	var connnections int
	switch topology {

	case "star":
		if pid == 0 { // Coordinator connects to all
			connnections = n - 1
			log.Println("Connecting to", connnections, "peers")
			GossipConnect(peers.SortByPeerID(), connnections)
		} else { // Non-coordinator waits for coordinator connection
			connnections = 1
		}

	case "gossip":
		connnections = k
		log.Println("Connecting to", connnections, "peers")
		GossipConnect(peers.Shuffle(), connnections)

	case "full":
		connnections = n - 1
		log.Println("Connecting to", connnections, "peers")
		//gossip.ConnectionSleepInterval = 10 * time.Second
		GossipConnect(peers.SortByPeerID(), connnections)
		//// Adaptation of gossip transport for full-connectivity
		//		gtransport.Validator = &validator.FullBuilder{pid, false}
	}

	gossipSetupTimeout = 20 * time.Second
	count = <-GossipWait(connnections).Done
	log.Println("Connected to", count, "peers after",
		time.Now().Sub(timestamp))

	if count < connnections {
		panic(fmt.Errorf("ERROR: expected %d neighbors, connected to %d",
			connnections, count))
	}

	wconfig := workload.DefaultConfig()
	wconfig.Log = log
	wconfig.LogDirectory = logDirectory
	wconfig.RandomValuesBytesSize = size
	wconfig.WarmupValuesCount = 0 // Number of epochs, 0 means no warm-up
	wconfig.MaxEpoch = maxEpoch
	if err := wconfig.Validate(); err != nil {
		panic(err)
	}
	workload := workload.NewGenerator(pid, wconfig)
	config := tendermint.DefaultConfig()
	keys := DeterministicKeySet(eid, n)
	config.PrivateKeys = keys.PrivateKeys
	config.PublicKeys = keys.PublicKeys
	config.VerifySignatures = true
	config.Log = log
	config.StatsPublishingInterval = 5 * time.Second
	config.TimeoutSmallDelta = time.Duration(smallDelta) * time.Millisecond
	config.TimeoutBigDelta = time.Duration(bigDelta) * time.Millisecond
	config.Model = model
	config.FastAlterEnabled = fastOpt
	config.MaxEpochToStart = maxEpoch
	if randomSeed == 0 {
		randomSeed = eid
	}
	if numByzantines > 0 {
		byzantines := generateByzantines(numByzantines, n, randomSeed)
		if byzantines[pid] {
			config.Byzantines = byzantines
			log.Println("Byzantine process")
		}
	}
	config.ByzTime = byzTime
	config.ByzAttack = byzAttack
	config.ChunksNumber = chunksNumber
	process = tendermint.NewProcess(pid, n, config, gtransport, workload)
	log.Printf("Created Tendermint process in zone %v\n", zone)

	stopChan := make(chan struct{})
	go workload.ProduceValues(stopChan)

	timestamp = time.Now()
	process.Bootstrap()
	bduration := time.Now().Sub(timestamp)
	log.Println("Bootstraped process in", bduration)

	go statsRoutine()
	go process.MainLoop()
	workload.Run(time.Duration(maxDuration)*time.Second, stopChan)

	//	coolDownTime := time.Duration(coolTime) * time.Second
	//	workload.NoopRoutine(coolDownTime)
}

// This should check that in every zone we have at least one byzantine.

// Force the agent to exit after 'duration' seconds.
func forceExit(duration int) {
	start := time.Now()
	time.Sleep(time.Duration(duration) * time.Second)
	log.Println("forced exit after", time.Now().Sub(start))
	panic("Agent maximum duration reached")

}

func generateByzantines(f, n int, seed int64) map[int]bool {
	byzantines := getRandomByzantines(f, n, seed)
	return byzantines
}

func getRandomByzantines(f, n int, seed int64) map[int]bool {
	src := rand.NewSource(seed)
	randGen := rand.New(src)
	byzantines := make(map[int]bool, f)
	for len(byzantines) != f {
		id := randGen.Intn(n - 1)
		if id != 0 {
			byzantines[id] = true
		}
	}
	return byzantines
}
