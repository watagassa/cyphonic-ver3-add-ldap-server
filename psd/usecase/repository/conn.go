package repository

type ConnectionRepository interface {
	LocalAddr() string
	RemoteAddr() string
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}
