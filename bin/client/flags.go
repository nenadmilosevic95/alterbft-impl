package main

import (
	"flag"

	"dslab.inf.usi.ch/tendermint/net/gossip"
)

// Flags for agents

var advertiseProxy bool

var capacity int
var ibcap int
var obcap int
var fvcap int

var clientMode bool
var semanticFiltering bool

var msgLossRate float64

func init() {
	// Agent setup
	flag.BoolVar(&advertiseProxy, "proxy", false, "Advertise as a proxy for the agent's zone.")
	flag.BoolVar(&clientMode, "client", false, "Operate in client-mode.")
	flag.StringVar(&listenAddr, "l", "", "Host listen adddress in Multiaddr format")

	// Consensus setup
	flag.IntVar(&capacity, "cap", 1024, "Capacity of all storages.")
	flag.IntVar(&fvcap, "fvcap", 1024, "Full value storage capacity.")

	// Transport setup
	flag.IntVar(&k, "k", 0, "Target number of neighbors.")
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
	flag.IntVar(&gossip.LRUCacheSize, "gcache", 262144, "Gossip LRU cache size.")
	flag.BoolVar(&semanticFiltering, "sfilter", false, "Enables semantic filtering via gossip validator.")
	flag.Float64Var(&msgLossRate, "msgloss", 0.0, "Message loss rate.")
}
