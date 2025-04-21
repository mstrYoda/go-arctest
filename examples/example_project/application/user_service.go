package application

import (
	"errors"
	"fmt"

	"github.com/mstrYoda/go-arctest/examples/example_project/domain"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo domain.UserRepositoryInterface
}

// NewUserService creates a new UserService
func NewUserService(userRepo domain.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(id string) (*domain.User, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	return s.userRepo.FindByID(id)
}

// CreateUser creates a new user
func (s *UserService) CreateUser(username, email string) (*domain.User, error) {
	if username == "" || email == "" {
		return nil, errors.New("username and email are required")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.FindByUsername(username)
	if existingUser != nil {
		return nil, fmt.Errorf("user with username %s already exists", username)
	}

	// Create new user
	user := &domain.User{
		ID:       generateID(), // In a real app, this would be a proper ID generation
		Username: username,
		Email:    email,
	}

	err := s.userRepo.Save(user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

// generateID is a simple ID generator for this example
func generateID() string {
	return "user-123" // This would be a proper UUID in a real application
}
