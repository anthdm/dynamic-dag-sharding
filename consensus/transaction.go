package consensus

type Transaction interface {
	Hash() []byte
}
