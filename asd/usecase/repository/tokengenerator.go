package repository

type TokenGenerator interface {
	GenerateAccessToken([]byte) (string, error)
}
