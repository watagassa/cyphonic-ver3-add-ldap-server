package entity

// Account is to set a CYPHONIC account.
type Account struct {
	ID       int    `gorm:"id"`
	Email    string `gorm:"column:email"`
	Password string `gorm:"column:password"`
	State    int    `gorm:"column:state"`
}
