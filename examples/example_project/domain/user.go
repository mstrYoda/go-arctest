package domain

// User represents a user in the system
type User struct {
	ID       string
	Username string
	Email    string
}

// UserRepositoryInterface defines the operations for User persistence
type UserRepositoryInterface interface {
	FindByID(id string) (*User, error)
	FindByUsername(username string) (*User, error)
	Save(user *User) error
	Delete(id string) error
}
