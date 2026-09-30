package message

import "fmt"

// CustomError defines error responses in JSON format.
type CustomError struct {
	ErrorMessage string `json:"error"`
}

func (err *CustomError) Error() string {
	return fmt.Sprintf(err.ErrorMessage)
}
