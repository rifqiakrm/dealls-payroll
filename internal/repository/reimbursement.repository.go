package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// ReimbursementRepository defines the interface for reimbursement data operations.
//
//go:generate mockgen -source=reimbursement.repository.go -destination=../../tests/mocks/repository/mock_reimbursement_repository.go -package=mocks
type ReimbursementRepository interface {
	CreateReimbursement(reimbursement *domain.Reimbursement) error
	GetReimbursementByID(id uuid.UUID) (*domain.Reimbursement, error)
	GetReimbursementsByUserIDAndPeriod(userID uuid.UUID, startDate, endDate time.Time) ([]domain.Reimbursement, error)
	UpdateReimbursement(reimbursement *domain.Reimbursement) error
	UpdateReimbursementsTx(tx *gorm.DB, reimbursements []domain.Reimbursement) error
}

// ReimbursementGormRepository implements repository.ReimbursementRepository using GORM.
type ReimbursementGormRepository struct {
	db *gorm.DB
}

// NewReimbursementGormRepository creates a new ReimbursementGormRepository.
func NewReimbursementGormRepository(db *gorm.DB) ReimbursementRepository {
	return &ReimbursementGormRepository{db: db}
}

// CreateReimbursement creates a new reimbursement record in the database.
func (r *ReimbursementGormRepository) CreateReimbursement(reimbursement *domain.Reimbursement) error {
	return r.db.Create(reimbursement).Error
}

// GetReimbursementByID retrieves a reimbursement record by its ID.
func (r *ReimbursementGormRepository) GetReimbursementByID(id uuid.UUID) (*domain.Reimbursement, error) {
	var reimbursement domain.Reimbursement
	err := r.db.First(&reimbursement, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &reimbursement, err
}

// GetReimbursementsByUserIDAndPeriod retrieves reimbursement records for a user within a date range.
func (r *ReimbursementGormRepository) GetReimbursementsByUserIDAndPeriod(userID uuid.UUID, startDate, endDate time.Time) ([]domain.Reimbursement, error) {
	var reimbursements []domain.Reimbursement
	err := r.db.Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).Find(&reimbursements).Error
	return reimbursements, err
}

// UpdateReimbursement updates an existing reimbursement record in the database.
func (r *ReimbursementGormRepository) UpdateReimbursement(reimbursement *domain.Reimbursement) error {
	return r.db.Save(reimbursement).Error
}

// UpdateReimbursementsTx updates multiple reimbursement records within the given transaction.
func (r *ReimbursementGormRepository) UpdateReimbursementsTx(tx *gorm.DB, reimbursements []domain.Reimbursement) error {
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	for _, reimbursement := range reimbursements {
		if err := tx.Save(&reimbursement).Error; err != nil {
			return err
		}
	}
	return nil
}
