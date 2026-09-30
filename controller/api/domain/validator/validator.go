package validator

import (
	"fmt"

	"github.com/go-playground/validator"
)

// CustomValidator returns Validator.
type CustomValidator struct {
	Validator *validator.Validate
}

// Validate validates data.
func (cv *CustomValidator) Validate(i interface{}) error {
	return fmt.Errorf("%w", cv.Validator.Struct(i))
}
