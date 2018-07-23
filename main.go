package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/anthdm/avalanche/consensus"
)

func main() {
	sim := consensus.NewNetworkSimulation(10, 0)
	sim.Run()
}

func init() {
	log.SetFlags(0)
	rand.Seed(time.Now().UnixNano())
}
