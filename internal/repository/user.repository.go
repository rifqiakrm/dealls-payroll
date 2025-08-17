package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// UserRepository defines the interface for user data operations.
//
//go:generate mockgen -source=user.repository.go -destination=../../tests/mocks/repository/mock_user_repository.go -package=mocks
type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByUsername(username string) (*domain.User, error)
	GetUserByID(id uuid.UUID) (*domain.User, error)
}

// UserGormRepository implements repository.UserRepository using GORM.
type UserGormRepository struct {
	db *gorm.DB
}

// NewUserGormRepository creates a new UserGormRepository.
func NewUserGormRepository(db *gorm.DB) UserRepository {
	return &UserGormRepository{db: db}
}

// CreateUser creates a new user in the database.
func (r *UserGormRepository) CreateUser(user *domain.User) error {
	return r.db.Create(user).Error
}

// GetUserByUsername retrieves a user by their username.
func (r *UserGormRepository) GetUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // User not found
	}
	return &user, err
}

// GetUserByID retrieves a user by their ID.
func (r *UserGormRepository) GetUserByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // User not found
	}
	return &user, err
}
