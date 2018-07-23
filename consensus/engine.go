package consensus

import (
	"crypto/ecdsa"
	"encoding/hex"
	"log"
	"sync"
)

type TxStatus uint8

const (
	Undefined TxStatus = iota
	Invalid
	Valid
)

// Algortithm tuning parameters.
const (
	convictionTreshold = 0.75
	treshold           = 0.75
	maxEpoch           = 3
	samples            = 4
)

type TxState struct {
	tx Transaction

	responses []TxStatus
	epoch     uint64

	// Confidence for our statusses.
	statusConf map[TxStatus]int
	cnt        int

	status TxStatus

	// The last status decision that is been made in an epoch.
	lastStatus TxStatus
}

func newTxState(tx Transaction, s TxStatus) *TxState {
	conf := map[TxStatus]int{
		Valid:   0,
		Invalid: 0,
	}

	return &TxState{
		responses:  []TxStatus{},
		status:     s,
		statusConf: conf,
		tx:         tx,
	}
}

func (s *TxState) isFinal() bool {
	return s.epoch >= maxEpoch
}

// Advance to the next epoch implying a reset in responses and increment the epoch
// counter. Will return true if we reached the max number of epochs.
func (s *TxState) advance() bool {
	s.epoch++
	s.responses = []TxStatus{}

	return s.epoch == maxEpoch
}

// Increment the confidence for the given status and return its current confidence.
func (s *TxState) incrStatus(status TxStatus) int {
	s.statusConf[status]++
	return s.statusConf[status]
}

func (s *TxState) countResponses(st TxStatus) int {
	i := 0
	for _, status := range s.responses {
		if status == st {
			i++
		}
	}
	return i
}

type Engine struct {
	id    uint64
	msgCh chan<- Message

	// The shard number this engine belongs to.
	shard int

	lock    sync.RWMutex
	mempool map[string]*TxState
}

func NewEngine(id uint64, msgCh chan<- Message) *Engine {
	return &Engine{
		id:      id,
		msgCh:   msgCh,
		mempool: make(map[string]*TxState),
	}
}

func (e *Engine) HandleMessage(from uint64, msg Message) error {
	switch p := msg.Payload.(type) {
	case Query:
		return e.handleQuery(from, p)
	case Response:
		result, err := e.handleResponse(from, p)
		if err != nil {
			return err
		}

		if result != nil {
			// only node 0 will print something on the screen.
			if e.id == 0 && result.status == Valid {
				e.lock.Lock()
				delete(e.mempool, string(result.hash))
				e.lock.Unlock()
				log.Printf("[CONFIRMED]\tshard: %d\t%s", e.shard, hex.EncodeToString(result.hash))
			}
		}
	}
	return nil
}

func (e *Engine) handleQuery(from uint64, q Query) error {
	state := e.getState(q.Tx.Hash())
	if state == nil {
		e.lock.Lock()
		e.mempool[string(q.Tx.Hash())] = newTxState(q.Tx, q.Status)
		e.lock.Unlock()
		return e.sendQuery(q.Tx, q.Status)
	}
	return e.sendResponse(from, q.Tx.Hash(), state.status)
}

func (e *Engine) getState(hash []byte) *TxState {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.mempool[string(hash)]
}

type Result struct {
	hash   []byte
	status TxStatus
}

func (e *Engine) handleResponse(from uint64, r Response) (*Result, error) {
	state := e.getState(r.Hash)
	if state == nil {
		return nil, nil //fmt.Errorf("could not find state related to response for tx %v", r.Hash)
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	// Already made decision for this tx, return nil do nothing.
	if state.isFinal() {
		return nil, nil
	}

	state.responses = append(state.responses, r.Status)
	n := state.countResponses(r.Status)
	if float64(n) >= (treshold * samples) {
		conf := state.incrStatus(r.Status)
		ourConf := state.statusConf[state.status]

		if conf > ourConf {
			state.status = r.Status
			state.lastStatus = r.Status
		}

		if r.Status != state.lastStatus {
			state.lastStatus = r.Status
			state.cnt = 0
		} else {
			state.cnt++
			if float64(state.cnt) > (convictionTreshold * samples) {
				if state.advance() {
					return &Result{
						hash:   state.tx.Hash(),
						status: state.status,
					}, nil
				}
			}
		}
	}

	return nil, e.sendQuery(state.tx, state.status)
}

func (e *Engine) handleTransaction(tx Transaction) error {
	ourTx := tx.(*TX)

	if ourTx.Shard != e.shard {
		log.Printf("[RELAY]\t\tshard: %d\t%s", e.shard, hex.EncodeToString(tx.Hash()))
		e.msgCh <- Message{e.id, nil, tx}
		return nil
	}

	// TODO: simulate tx verification. For now just set valid.
	status := verifyTransaction(tx)

	if e.getState(tx.Hash()) == nil {
		e.lock.Lock()
		e.mempool[string(tx.Hash())] = newTxState(tx, status)
		e.lock.Unlock()
		return e.sendQuery(tx, status)
	}
	return nil
}

func (e *Engine) sendQuery(tx Transaction, status TxStatus) error {
	msg := Message{e.id, nil, Query{tx, status}}
	e.msgCh <- msg
	return nil
}

func (e *Engine) sendResponse(to uint64, hash []byte, status TxStatus) error {
	msg := Message{e.id, &to, Response{hash, status}}
	e.msgCh <- msg
	return nil
}

func verifyTransaction(tx Transaction) TxStatus {
	ourTx := tx.(*TX)
	if ecdsa.Verify(&ourTx.PublicKey, ourTx.Hash(), ourTx.R, ourTx.S) {
		return Valid
	}
	return Invalid
}
