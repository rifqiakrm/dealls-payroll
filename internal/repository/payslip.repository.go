package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// PayslipRepository defines the interface for payslip data operations.
//
//go:generate mockgen -source=payslip.repository.go -destination=../../tests/mocks/repository/mock_payslip_repository.go -package=mocks
type PayslipRepository interface {
	CreatePayslip(payslip *domain.Payslip) error
	GetPayslipByID(id uuid.UUID) (*domain.Payslip, error)
	GetPayslipByUserIDAndPeriodID(userID, periodID uuid.UUID) (*domain.Payslip, error)
	GetAllPayslipsByPeriodID(periodID uuid.UUID) ([]domain.Payslip, error)
	CreatePayslipTx(tx *gorm.DB, payslip *domain.Payslip) error
}

// PayslipGormRepository implements repository.PayslipRepository using GORM.
type PayslipGormRepository struct {
	db *gorm.DB
}

// NewPayslipGormRepository creates a new PayslipGormRepository.
func NewPayslipGormRepository(db *gorm.DB) PayslipRepository {
	return &PayslipGormRepository{db: db}
}

// CreatePayslip creates a new payslip record in the database.
func (r *PayslipGormRepository) CreatePayslip(payslip *domain.Payslip) error {
	return r.db.Create(payslip).Error
}

// GetPayslipByID retrieves a payslip record by its ID.
func (r *PayslipGormRepository) GetPayslipByID(id uuid.UUID) (*domain.Payslip, error) {
	var payslip domain.Payslip
	err := r.db.First(&payslip, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &payslip, err
}

// GetPayslipByUserIDAndPeriodID retrieves a payslip record by user ID and payroll period ID.
func (r *PayslipGormRepository) GetPayslipByUserIDAndPeriodID(userID, periodID uuid.UUID) (*domain.Payslip, error) {
	var payslip domain.Payslip
	err := r.db.
		Where("user_id = ? AND payroll_period_id = ?", userID, periodID).
		First(&payslip).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &payslip, err
}

// GetAllPayslipsByPeriodID retrieves all payslip records for a given payroll period ID.
func (r *PayslipGormRepository) GetAllPayslipsByPeriodID(periodID uuid.UUID) ([]domain.Payslip, error) {
	var payslips []domain.Payslip
	err := r.db.
		Where("payroll_period_id = ?", periodID).
		Find(&payslips).Error
	return payslips, err
}

// CreatePayslipTx inserts a new payslip record within the given transaction.
func (r *PayslipGormRepository) CreatePayslipTx(tx *gorm.DB, payslip *domain.Payslip) error {
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	return tx.Create(payslip).Error
}
