package consensus

// Message represents the top level message that is being sent is the gossip
// consensus protocol.
type Message struct {
	// Id of the sender.
	Origin uint64
	// Id of the receiver. If nil the message will be broadcasted to a sample
	// of the network.
	To *uint64
	// Payload that is carried with the message.
	Payload interface{}
}

type Query struct {
	Tx     Transaction
	Status TxStatus
}

type Response struct {
	Hash   []byte
	Status TxStatus
}
