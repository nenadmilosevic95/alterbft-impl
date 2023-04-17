package main

import (
	"flag"
)

// Flags for clients

var rate float64
var size int
var duration int
var maxDuration int
var perfInterval int
var logDirectory string

func init() {
	flag.Float64Var(&rate, "rate", 1, "Rate of proposals (values/sec).")
	flag.IntVar(&size, "s", 1024, "Size of proposed values.")
	flag.IntVar(&duration, "d", 15, "Duration of the experiment in seconds.")
	flag.IntVar(&maxDuration, "dmax", 30, "Maximum duration of the experiment in seconds.")
	flag.IntVar(&perfInterval, "p", 5, "Performance stats interval in seconds.")
	flag.StringVar(&logDirectory, "dir", ".", "Directory for writting log files, it must exist.")
}
