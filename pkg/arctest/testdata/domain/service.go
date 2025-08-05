package domain

// UserService provides domain logic for users
type UserService struct {
	repo UserRepositoryInterface
}

// NewUserService creates a new UserService with repository dependency
func NewUserService(repo UserRepositoryInterface) *UserService {
	return &UserService{
		repo: repo,
	}
}

// FindUser finds a user by ID
func (s *UserService) FindUser(id string) (*UserEntity, error) {
	return s.repo.FindByID(id)
}

// SaveUser saves a user
func (s *UserService) SaveUser(user *UserEntity) error {
	return s.repo.Save(user)
}
