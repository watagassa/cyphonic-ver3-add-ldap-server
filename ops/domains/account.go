// Package domains contains the database informations.
package domains

// Accounts is to set a CYPHONIC account.
type Accounts struct {
	ID       int    `gorm:"id"`
	Email    string `gorm:"column:email"`
	Password string `gorm:"column:password"`
	State    int    `gorm:"column:state"`
}
