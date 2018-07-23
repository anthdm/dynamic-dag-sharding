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

	Shard int

	// Signature values
	R *big.Int
	S *big.Int
}

func randomTx() *TX {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	tx := &TX{
		Nonce: rand.Uint64(),
		Data:  rand.Uint64(),
		Shard: rand.Intn(len([]int{0, 1})),
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

var txInterval = 3 * time.Millisecond

type NetworkSimulation struct {
	engines map[uint64]*Engine
	msgCh   chan Message

	latency time.Duration
}

func NewNetworkSimulation(n int, lat int) *NetworkSimulation {
	engines := make(map[uint64]*Engine)
	// TODO: Make sure this channel will not block all goroutines! 128 buffer
	// per engine should be fine though.
	//msgCh := make(chan Message, n*512)
	msgCh := make(chan Message)

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
	var (
		txTimer   = time.NewTimer(txInterval)
		quitTimer = time.NewTimer(10 * time.Second)
	)

free:
	for {
		select {
		case <-txTimer.C:
			tx := randomTx()
			log.Printf("[NEW]\t\tshard: %d\t%s", tx.Shard, hex.EncodeToString(tx.Hash()))

			e := sim.engines[uint64(rand.Intn(len(sim.engines)))]
			go func(e *Engine) {
				if err := e.handleTransaction(tx); err != nil {
					log.Fatal(err)
				}
			}(e)
			txTimer.Reset(txInterval)
		case <-quitTimer.C:
			break free
		case msg := <-sim.msgCh:
			switch p := msg.Payload.(type) {
			case *Transaction:
				for _, e := range sim.sampleEngines(samples) {
					go func(e *Engine) {
						time.Sleep(sim.latency)
						if err := e.handleTransaction(*p); err != nil {
							log.Fatal(err)
						}
					}(e)
				}
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
