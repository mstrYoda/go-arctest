package presentation

import (
	"github.com/mstrYoda/go-arctest/pkg/arctest/testdata/application"
	"github.com/mstrYoda/go-arctest/pkg/arctest/testdata/domain"
)

// UserController handles user-related requests
type UserController struct {
	appService *application.UserApplicationService
}

// NewUserController creates a new UserController
func NewUserController(appService *application.UserApplicationService) *UserController {
	return &UserController{
		appService: appService,
	}
}

// GetUser gets a user by ID
func (c *UserController) GetUser(id string) (*domain.UserEntity, error) {
	return c.appService.GetUser(id)
}

// CreateUser creates a new user
func (c *UserController) CreateUser(id, name, email string) error {
	user := &domain.UserEntity{
		ID:    id,
		Name:  name,
		Email: email,
	}

	return c.appService.SaveUser(user)
}
