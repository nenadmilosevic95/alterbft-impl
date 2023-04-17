package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

var cpuprofile string
var memprofile string

func init() {
	// CPU and memory profiles
	flag.StringVar(&cpuprofile, "cpuf", "", "write cpu profile to `file`")
	flag.StringVar(&memprofile, "memf", "", "write mem profile to `file`")
}

func profilerLoop() {
	var memf *os.File
	var cpuf *os.File
	var err error
	if len(cpuprofile) > 0 {
		cpuf, err = os.Create(fmt.Sprint(cpuprofile, ".", pid))
		if err != nil {
			panic(err)
		}
		if err := pprof.StartCPUProfile(cpuf); err != nil {
			panic(err)
		}
	}
	if len(memprofile) > 0 {
		memf, err = os.Create(fmt.Sprint(memprofile, ".", pid))
		if err != nil {
			panic(err)
		}
	}
	time.Sleep(time.Duration(duration) * time.Second)
	if cpuf != nil {
		pprof.StopCPUProfile()
		cpuf.Close()
	}
	if memf != nil {
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(memf); err != nil {
			panic(err)
		}
	}
}
