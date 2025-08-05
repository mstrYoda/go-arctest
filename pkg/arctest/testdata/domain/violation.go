package domain

import (
	// This import violates domain layer rules
	"github.com/mstrYoda/go-arctest/pkg/arctest/testdata/infrastructure"
)

// ViolatingService incorrectly depends on infrastructure
type ViolatingService struct {
	repo *infrastructure.UserRepository
}

// NewViolatingService creates a new service that violates architecture rules
func NewViolatingService(repo *infrastructure.UserRepository) *ViolatingService {
	return &ViolatingService{
		repo: repo,
	}
}

// FindUser finds a user by ID
func (s *ViolatingService) FindUser(id string) (*UserEntity, error) {
	return s.repo.FindByID(id)
}
