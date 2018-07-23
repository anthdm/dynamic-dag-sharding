package consensus

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log"
	"math/big"
	"math/rand"
	"time"
)

type TX struct {
	Nonce     uint64
	Data      uint64
	PublicKey ecdsa.PublicKey

	// Signature values
	R *big.Int
	S *big.Int
}

func RandomTx() *TX {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	tx := &TX{
		Nonce: rand.Uint64(),
		Data:  rand.Uint64(),
	}
	tx.Sign(priv)
	return tx
}

func (tx *TX) Sign(priv *ecdsa.PrivateKey) {
	hash := tx.Hash()
	r, s, err := ecdsa.Sign(crand.Reader, priv, hash)
	// for the sake of the POC just fatal here.
	if err != nil {
		log.Fatal(err)
	}
	tx.R = r
	tx.S = s
	tx.PublicKey = priv.PublicKey
}

func (tx *TX) Serialize() []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[:8], tx.Nonce)
	binary.LittleEndian.PutUint64(buf[8:], tx.Data)
	return buf
}

func (tx *TX) Hash() []byte {
	sha := sha256.New()
	sha.Write(tx.Serialize())
	return sha.Sum(nil)
}

var txInterval = 50 * time.Millisecond

type NetworkSimulation struct {
	engines map[uint64]*Engine
	msgCh   chan Message

	latency time.Duration
}

func NewNetworkSimulation(n int, lat int) *NetworkSimulation {
	engines := make(map[uint64]*Engine)
	// TODO: Make sure this channel will not block all goroutines! 128 buffer
	// per engine should be fine though.
	msgCh := make(chan Message, n*512)

	for i := 0; i < n; i++ {
		id := uint64(i)
		engines[id] = NewEngine(id, msgCh)
	}

	return &NetworkSimulation{
		engines: engines,
		msgCh:   msgCh,
		latency: time.Duration(lat) * time.Millisecond,
	}
}

func (sim *NetworkSimulation) Run() {
	txTimer := time.NewTimer(txInterval)
	for {
		select {
		case <-txTimer.C:
			e := sim.engines[uint64(rand.Intn(len(sim.engines)))]
			tx := RandomTx()
			log.Printf("sending tx %s into the network", hex.EncodeToString(tx.Hash()))

			if err := e.handleTransaction(tx); err != nil {
				log.Fatal(err)
			}
			txTimer.Reset(txInterval)
		case msg := <-sim.msgCh:
			switch msg.Payload.(type) {
			case Query:
				for _, e := range sim.sampleEngines(samples) {
					go func(e *Engine) {
						time.Sleep(sim.latency)
						if err := e.HandleMessage(msg.Origin, msg); err != nil {
							log.Fatal(err)
						}
					}(e)
				}
			case Response:
				e, ok := sim.engines[*msg.To]
				if !ok {
					log.Fatalf("cannot send response to unknown engine id (%d)", msg.To)
				}
				go func(e *Engine) {
					time.Sleep(sim.latency)
					if err := e.HandleMessage(msg.Origin, msg); err != nil {
						log.Fatal(err)
					}
				}(e)
			}
		}
	}
}

func (sim *NetworkSimulation) sampleEngines(n int) []*Engine {
	if n > len(sim.engines) {
		panic("cannot sample more engines then available")
	}
	s := make([]*Engine, len(sim.engines))
	for i, e := range sim.engines {
		s[i] = e
	}

	engines := make([]*Engine, n)
	for i := 0; i < n; i++ {
		engines[i] = s[rand.Intn(len(s))]
	}

	return engines
}
