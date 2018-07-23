package main

import (
	"log"

	"github.com/anthdm/avalanche/consensus"
)

func main() {
	sim := consensus.NewNetworkSimulation(20, 80)
	sim.Run()
}

func init() {
	log.SetFlags(0)
}
