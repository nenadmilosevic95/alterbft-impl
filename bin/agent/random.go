package main

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
)

func setupRandomGenerator() {
	if randomSeed == 0 {
		randomSeed = eid
		log.Println("random seed:", randomSeed, "(experiment ID)")
	} else {
		log.Println("random seed:", randomSeed, "(set by parameter)")
	}

	// Time is not a good random source, so hash the randomSeed
	hash := fnv.New64a()
	array := make([]byte, 16)
	binary.LittleEndian.PutUint64(array, uint64(randomSeed))
	hash.Write(array)

	// Generate some dispersion using process id and n
	deterministicShift := uint64(pid * n * 200)
	rand.Seed(int64(hash.Sum64() + deterministicShift))
}
