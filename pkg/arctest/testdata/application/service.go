package application

import (
	"github.com/mstrYoda/go-arctest/pkg/arctest/testdata/domain"
	"github.com/mstrYoda/go-arctest/pkg/arctest/testdata/infrastructure"
)

// UserApplicationService has methods for user operations
type UserApplicationService struct {
	userService *domain.UserService
}

// NewUserApplicationService creates a new UserApplicationService with proper interface
func NewUserApplicationService(repo domain.UserRepositoryInterface) *UserApplicationService {
	return &UserApplicationService{
		userService: domain.NewUserService(repo),
	}
}

// This constructor violates our architecture rules by using concrete implementation
// instead of interface
func NewUserApplicationServiceWithViolation(repo *infrastructure.UserRepository) *UserApplicationService {
	return &UserApplicationService{
		userService: domain.NewUserService(repo),
	}
}

// GetUser returns a user by ID
func (s *UserApplicationService) GetUser(id string) (*domain.UserEntity, error) {
	return s.userService.FindUser(id)
}

// SaveUser saves a user
func (s *UserApplicationService) SaveUser(user *domain.UserEntity) error {
	return s.userService.SaveUser(user)
}
