// Package domain is the entity of the data model.
package domain

import "time"

// Account is to set a CYPHONIC account.
type Account struct {
	ID        int       `json:"id" validate:"required"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	State     int       `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
