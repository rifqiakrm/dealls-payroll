package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// PayrollPeriodRepository defines the interface for payroll period data operations.
//
//go:generate mockgen -source=payroll_period.repository.go -destination=../../tests/mocks/repository/mock_payroll_period_repository.go -package=mocks
type PayrollPeriodRepository interface {
	CreatePayrollPeriod(period *domain.PayrollPeriod) error
	GetPayrollPeriodByID(id uuid.UUID) (*domain.PayrollPeriod, error)
	GetActivePayrollPeriod() (*domain.PayrollPeriod, error)
	MarkPayrollPeriodAsProcessed(id uuid.UUID) error
	GetAllPayrollPeriods() ([]domain.PayrollPeriod, error)
	GetPayrollPeriodByDates(startDate, endDate time.Time) (*domain.PayrollPeriod, error)
	MarkPayrollPeriodAsProcessedTx(tx *gorm.DB, periodID uuid.UUID) error
	GetOverlappingPayrollPeriods(startDate, endDate time.Time) ([]domain.PayrollPeriod, error)
}

// PayrollPeriodGormRepository implements repository.PayrollPeriodRepository using GORM.
type PayrollPeriodGormRepository struct {
	db *gorm.DB
}

// NewPayrollPeriodGormRepository creates a new PayrollPeriodGormRepository.
func NewPayrollPeriodGormRepository(db *gorm.DB) PayrollPeriodRepository {
	return &PayrollPeriodGormRepository{db: db}
}

// CreatePayrollPeriod creates a new payroll period in the database.
func (r *PayrollPeriodGormRepository) CreatePayrollPeriod(period *domain.PayrollPeriod) error {
	return r.db.Create(period).Error
}

// GetPayrollPeriodByID retrieves a payroll period by its ID.
func (r *PayrollPeriodGormRepository) GetPayrollPeriodByID(id uuid.UUID) (*domain.PayrollPeriod, error) {
	var period domain.PayrollPeriod
	err := r.db.First(&period, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &period, err
}

// GetActivePayrollPeriod retrieves the currently active (not processed) payroll period.
func (r *PayrollPeriodGormRepository) GetActivePayrollPeriod() (*domain.PayrollPeriod, error) {
	var period domain.PayrollPeriod
	err := r.db.Where("is_processed = ?", false).Order("start_date ASC").First(&period).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &period, err
}

// MarkPayrollPeriodAsProcessed updates a payroll period's status to processed.
func (r *PayrollPeriodGormRepository) MarkPayrollPeriodAsProcessed(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&domain.PayrollPeriod{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_processed": true,
		"processed_at": &now,
	}).Error
}

// GetAllPayrollPeriods retrieves all payroll periods.
func (r *PayrollPeriodGormRepository) GetAllPayrollPeriods() ([]domain.PayrollPeriod, error) {
	var periods []domain.PayrollPeriod
	err := r.db.Order("start_date DESC").Find(&periods).Error
	return periods, err
}

// GetPayrollPeriodByDates retrieves a payroll period by its start and end dates.
func (r *PayrollPeriodGormRepository) GetPayrollPeriodByDates(startDate, endDate time.Time) (*domain.PayrollPeriod, error) {
	var period domain.PayrollPeriod
	err := r.db.Where("start_date = ? AND end_date = ?", startDate, endDate).First(&period).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &period, err
}

// MarkPayrollPeriodAsProcessedTx marks a payroll period as processed within a transaction.
func (r *PayrollPeriodGormRepository) MarkPayrollPeriodAsProcessedTx(tx *gorm.DB, periodID uuid.UUID) error {
	result := tx.Model(&domain.PayrollPeriod{}).
		Where("id = ? AND is_processed = ?", periodID, false).
		Updates(map[string]interface{}{
			"is_processed": true,
			"processed_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark payroll period as processed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no payroll period updated, maybe already processed or not found")
	}

	return nil
}

// GetOverlappingPayrollPeriods retrieves payroll periods that overlap with the given date range.
// Overlap means: (period.StartDate <= endDate) AND (period.EndDate >= startDate).
func (r *PayrollPeriodGormRepository) GetOverlappingPayrollPeriods(startDate, endDate time.Time) ([]domain.PayrollPeriod, error) {
	var periods []domain.PayrollPeriod

	err := r.db.
		Where("start_date <= ? AND end_date >= ?", endDate, startDate).
		Find(&periods).Error

	if err != nil {
		return nil, err
	}

	return periods, nil
}
