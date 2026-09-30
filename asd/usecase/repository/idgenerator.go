package repository

type IDGenerator interface {
	GenerateNodeID(string, []byte) ([]byte, error)
}
