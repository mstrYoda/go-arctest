package infrastructure

import (
	"errors"
	"sync"

	"github.com/mstrYoda/go-arctest/examples/example_project/domain"
)

// UserRepository implements the domain.UserRepositoryInterface
type UserRepository struct {
	users map[string]*domain.User
	mu    sync.RWMutex
}

// NewUserRepository creates a new UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*domain.User),
	}
}

// FindByID retrieves a user by their ID
func (r *UserRepository) FindByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, found := r.users[id]
	if !found {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// FindByUsername retrieves a user by their username
func (r *UserRepository) FindByUsername(username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// Save stores a user in the repository
func (r *UserRepository) Save(user *domain.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	return nil
}

// Delete removes a user from the repository
func (r *UserRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, found := r.users[id]; !found {
		return errors.New("user not found")
	}

	delete(r.users, id)
	return nil
}
