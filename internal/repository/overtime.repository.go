package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// OvertimeRepository defines the interface for overtime data operations.
//
//go:generate mockgen -source=overtime.repository.go -destination=../../tests/mocks/repository/mock_overtime_repository.go -package=mocks
type OvertimeRepository interface {
	CreateOvertime(overtime *domain.Overtime) (*domain.Overtime, error)
	GetOvertimeByID(id uuid.UUID) (*domain.Overtime, error)
	GetOvertimeByUserIDAndDate(userID uuid.UUID, date time.Time) ([]domain.Overtime, error)
	GetOvertimesByUserIDAndPeriod(userID uuid.UUID, startDate, endDate time.Time) ([]domain.Overtime, error)
	GetOvertimesByUserIDAndPayrollPeriodID(userID uuid.UUID, payrollPeriodID uuid.UUID) ([]*domain.Overtime, error)
	UpdateOvertime(overtime *domain.Overtime) error
	UpdateOvertimesTx(tx *gorm.DB, overtimes []domain.Overtime) error
}

// OvertimeGormRepository implements repository.OvertimeRepository using GORM.
type OvertimeGormRepository struct {
	db *gorm.DB
}

// NewOvertimeGormRepository creates a new OvertimeGormRepository.
func NewOvertimeGormRepository(db *gorm.DB) OvertimeRepository {
	return &OvertimeGormRepository{db: db}
}

// CreateOvertime creates a new overtime record in the database.
func (r *OvertimeGormRepository) CreateOvertime(overtime *domain.Overtime) (*domain.Overtime, error) {
	return overtime, r.db.Create(overtime).Error
}

// GetOvertimeByID retrieves an overtime record by its ID.
func (r *OvertimeGormRepository) GetOvertimeByID(id uuid.UUID) (*domain.Overtime, error) {
	var overtime domain.Overtime
	err := r.db.First(&overtime, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &overtime, err
}

// GetOvertimeByUserIDAndDate retrieves overtime records for a user on a specific date.
func (r *OvertimeGormRepository) GetOvertimeByUserIDAndDate(userID uuid.UUID, date time.Time) ([]domain.Overtime, error) {
	var overtimes []domain.Overtime
	err := r.db.Where("user_id = ? AND date = ?", userID, date.Format("2006-01-02")).Find(&overtimes).Error
	return overtimes, err
}

// GetOvertimesByUserIDAndPeriod retrieves overtime records for a user within a date range.
func (r *OvertimeGormRepository) GetOvertimesByUserIDAndPeriod(userID uuid.UUID, startDate, endDate time.Time) ([]domain.Overtime, error) {
	var overtimes []domain.Overtime
	err := r.db.Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).Find(&overtimes).Error
	return overtimes, err
}

// GetOvertimesByUserIDAndPayrollPeriodID retrieves overtime records for a user by payroll period ID.
func (r *OvertimeGormRepository) GetOvertimesByUserIDAndPayrollPeriodID(userID uuid.UUID, payrollPeriodID uuid.UUID) ([]*domain.Overtime, error) {
	var overtimes []*domain.Overtime
	err := r.db.Where("user_id = ? AND payroll_period_id = ?", userID, payrollPeriodID).Find(&overtimes).Error
	return overtimes, err
}

// UpdateOvertime updates an existing overtime record in the database.
func (r *OvertimeGormRepository) UpdateOvertime(overtime *domain.Overtime) error {
	return r.db.Save(overtime).Error
}

// UpdateOvertimesTx updates multiple overtime records within the given transaction.
func (r *OvertimeGormRepository) UpdateOvertimesTx(tx *gorm.DB, overtimes []domain.Overtime) error {
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	for _, overtime := range overtimes {
		if err := tx.Save(&overtime).Error; err != nil {
			return err
		}
	}
	return nil
}
