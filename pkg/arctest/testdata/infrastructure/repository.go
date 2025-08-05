package infrastructure

import (
	"errors"
	"sync"

	"github.com/mstrYoda/go-arctest/pkg/arctest/testdata/domain"
)

// UserRepository implements the domain.UserRepositoryInterface
type UserRepository struct {
	users map[string]*domain.UserEntity
	mu    sync.RWMutex
}

// NewUserRepository creates a new UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*domain.UserEntity),
	}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id string) (*domain.UserEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// Save saves a user
func (r *UserRepository) Save(user *domain.UserEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	return nil
}

// ProductRepository does NOT implement the domain.ProductRepositoryInterface
// It's deliberately missing Save method to test violation detection
type ProductRepository struct {
	products map[string]*domain.ProductEntity
	mu       sync.RWMutex
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		products: make(map[string]*domain.ProductEntity),
	}
}

// FindByID finds a product by ID
func (r *ProductRepository) FindByID(id string) (*domain.ProductEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, ok := r.products[id]
	if !ok {
		return nil, errors.New("product not found")
	}

	return product, nil
}

// We intentionally omit Save method to test violation detection
