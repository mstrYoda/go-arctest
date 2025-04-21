package domain

import (
	"errors"

	"github.com/mstrYoda/go-arctest/examples/example_project/utils"
)

// This file intentionally violates the architecture by importing the utils package
// from the domain layer, which should not depend on it.

// UserService should be in the application layer, not the domain layer
// This is violating the architecture on purpose for testing
type UserServiceWithLogger struct {
	logger *utils.Logger
}

// NewUserServiceWithLogger creates a new UserServiceWithLogger
func NewUserServiceWithLogger() *UserServiceWithLogger {
	return &UserServiceWithLogger{
		logger: utils.NewLogger("UserService"),
	}
}

// CreateUserWithLogging creates a new user with logging
func (s *UserServiceWithLogger) CreateUserWithLogging(username, email string) (*User, error) {
	s.logger.Log("Attempting to create a new user")

	if username == "" || email == "" {
		err := errors.New("username and email are required")
		s.logger.LogError(err)
		return nil, err
	}

	user := &User{
		ID:       "user-" + username,
		Username: username,
		Email:    email,
	}

	s.logger.Log("User created successfully")
	return user, nil
}
