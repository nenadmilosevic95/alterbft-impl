package main

import (
	"fmt"
	"math/rand"
	"time"

	"dslab.inf.usi.ch/tendermint/net"
)

var rate float64
var size int
var duration int
var maxDuration int
var perfInterval int

var proposals []net.Decision
var pendingProposals = make(map[uint64]*net.Decision)

var deliveryQueue = make(chan *net.Decision, 128)

func clientLoop() {
	// Avoid the experiment to run forever
	go forceExit(maxDuration)

	rand.Seed(time.Now().UnixNano())
	log.Println("number of processes:", n)
	log.Println("inter-processes connections:", k)
	log.Println("size of proposed values:", size)

	perfDone := make(chan struct{})
	perfQueue := make(chan *Record, 100)
	go perfLoop(perfQueue, perfDone, perfInterval)
	log.Println("performance stats interval:", perfInterval)

	go deliveryLoop()

	// Proposer main loop
	log.Println("experiment duration:", duration)
	durationTS := time.Now().Add(time.Duration(duration) * time.Second)

	var tickInterval = time.Duration(float64(time.Second) / rate)
	log.Printf("Rate: %.1f values/s, proposals every %v\n",
		rate, tickInterval)

	// First submission
	active := true
	submitNextProposal()
	tick := time.Tick(tickInterval)

	for active || len(pendingProposals) > 0 {
		var decision *net.Decision
		select {
		case decision = <-deliveryQueue:
			proposal := pendingProposals[decision.ValueID]
			if proposal != nil {
				delete(pendingProposals, decision.ValueID)
				latency := decision.Timestamp.Sub(
					proposal.Timestamp)
				perfQueue <- &Record{
					Timestamp: decision.Timestamp,
					Latency:   latency,
				}

				//	if debug {
				//log.Println("decided ",
				//	decision.ValueID,
				//	"+", len(proposal.Value),
				//	"bytes in", latency)
				log.Println("decided", decision.ValueID, int64(latency/time.Millisecond), decision.Instance)
				//}

			}
		case now := <-tick:
			if now.Before(durationTS) {
				submitNextProposal()
			} else {
				active = false
			}
		}
	}

	log.Println("Current pending proposals:", len(pendingProposals))
	log.Println("Maximum pending proposals:", len(proposals))

	// Wait until performance records are parsed
	close(perfQueue)
	<-perfDone
}

func deliveryLoop() {
	for {
		decision, err := client.Decide()
		if err != nil {
			panic(fmt.Errorf("decide error: %v", err))
		}

		deliveryQueue <- decision

		if debug {
			//log.Printf("decision %v instance %v",
			//	decision.ValueID, decision.Instance)
		}
	}
}

func submitNextProposal() {
	var proposal *net.Decision
	for i := range proposals {
		if pendingProposals[proposals[i].ValueID] == nil {
			proposal = &proposals[i]
			break
		}
	}

	if proposal == nil {
		proposals = append(proposals, net.Decision{
			Value: make([]byte, size),
		})
		proposal = &proposals[len(proposals)-1]
	}

	rand.Read(proposal.Value)
	proposal.ValueID = net.ValueID(proposal.Value)
	pendingProposals[proposal.ValueID] = proposal
	proposal.Timestamp = time.Now()

	// Encapsulate value in a proposal and send it
	err := client.Propose(proposal.Value)
	if err != nil {
		panic(fmt.Errorf("propose error: %v", err))

	}

	if debug {
		log.Println("proposed", proposal.ValueID)
	}
}

// Force the proposer to exit after 'duration' seconds.
func forceExit(duration int) {
	start := time.Now()
	time.Sleep(time.Duration(duration) * time.Second)
	log.Println("forced exit after: ", time.Now().Sub(start))
	panic("Client agent maximum duration reached")

}

// Summary of performance data returned as a string.
func perfSummary(p *Perf) string {
	s := fmt.Sprintf("%d values", p.Values)
	s += fmt.Sprintf(", %.3f values/s", p.ValuesPerSec())
	s += fmt.Sprintf(", latency: %v +- %v",
		p.LatencyMean(), p.LatencyStdev())
	return s
}

// Performance loop reads performance records and computes performance data.
// Every 'interval' second, the current performance data summary is printed.
func perfLoop(queue chan *Record, done chan struct{}, interval int) {
	var record *Record
	var perf *Perf = &Perf{}
	var iperf *Perf = &Perf{}
	start := time.Now()
	ticker := time.Tick(time.Duration(interval) * time.Second)

Loop:
	for {
		select {
		case record = <-queue:
			if record == nil {
				break Loop
			}
		case now := <-ticker:
			fmt.Println("---------------------------------------------------------------------------")
			fmt.Println(now.Sub(start), "\t", perfSummary(perf))
			fmt.Println("   interval\t", perfSummary(iperf))
			iperf.Reset()
			continue
		}

		perf.Add(record)
		iperf.Add(record)
	}

	// Final performance results:
	duration := time.Now().Sub(start)
	fmt.Println("---------------------------------------------------------------------------")
	fmt.Println(duration, "\t", perfSummary(perf))
	fmt.Println("---------------------------------------------------------------------------")
	fmt.Printf("%d\t%.3f\t%d\t%.3f\t%.3f\t%.3f\n", n, duration.Seconds(),
		perf.Values, perf.ValuesPerSec(),
		perf.LatencyMean().Seconds()*1000.0,
		perf.LatencyStdev().Seconds()*1000.0)
	close(done)
}
