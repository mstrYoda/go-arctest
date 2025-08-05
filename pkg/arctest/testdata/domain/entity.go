package domain

// UserEntity represents a user in the domain
type UserEntity struct {
	ID    string
	Name  string
	Email string
}

// UserRepositoryInterface defines the interface for user repository
type UserRepositoryInterface interface {
	FindByID(id string) (*UserEntity, error)
	Save(user *UserEntity) error
}

// ProductEntity represents a product in the domain
type ProductEntity struct {
	ID       string
	Name     string
	Price    float64
	Quantity int
}

// ProductRepositoryInterface defines the interface for product repository
type ProductRepositoryInterface interface {
	FindByID(id string) (*ProductEntity, error)
	Save(product *ProductEntity) error
}
