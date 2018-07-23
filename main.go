package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/anthdm/dynamic-dag-sharding/consensus"
)

func main() {
	sim := consensus.NewNetworkSimulation(16, 100) // zero network latency for now.
	sim.Run()
}

func init() {
	log.SetFlags(0)
	rand.Seed(time.Now().UnixNano())
}
