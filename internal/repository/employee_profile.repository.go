package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// EmployeeProfileRepository defines the interface for employee profile data operations.
//
//go:generate mockgen -source=employee_profile.repository.go -destination=../../tests/mocks/repository/mock_employee_profile_repository.go -package=mocks
type EmployeeProfileRepository interface {
	CreateEmployeeProfile(profile *domain.EmployeeProfile) error
	GetEmployeeProfileByUserID(userID uuid.UUID) (*domain.EmployeeProfile, error)
	GetAllEmployeeProfiles() ([]domain.EmployeeProfile, error)
}

// EmployeeProfileGormRepository implements repository.EmployeeProfileRepository using GORM.
type EmployeeProfileGormRepository struct {
	db *gorm.DB
}

// NewEmployeeProfileGormRepository creates a new EmployeeProfileGormRepository.
func NewEmployeeProfileGormRepository(db *gorm.DB) EmployeeProfileRepository {
	return &EmployeeProfileGormRepository{db: db}
}

// CreateEmployeeProfile creates a new employee profile in the database.
func (r *EmployeeProfileGormRepository) CreateEmployeeProfile(profile *domain.EmployeeProfile) error {
	return r.db.Create(profile).Error
}

// GetEmployeeProfileByUserID retrieves an employee profile by user ID.
func (r *EmployeeProfileGormRepository) GetEmployeeProfileByUserID(userID uuid.UUID) (*domain.EmployeeProfile, error) {
	var profile domain.EmployeeProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &profile, err
}

// GetAllEmployeeProfiles retrieves all employee profiles.
func (r *EmployeeProfileGormRepository) GetAllEmployeeProfiles() ([]domain.EmployeeProfile, error) {
	var profiles []domain.EmployeeProfile
	err := r.db.Find(&profiles).Error
	return profiles, err
}
