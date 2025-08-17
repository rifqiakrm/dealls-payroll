package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// AttendanceRepository defines the interface for attendance data operations.
//
//go:generate mockgen -source=attendance.repository.go -destination=../../tests/mocks/repository/mock_attendance_repository.go -package=mocks
type AttendanceRepository interface {
	CreateAttendance(attendance *domain.Attendance) error
	GetAttendanceByID(id uuid.UUID) (*domain.Attendance, error)
	GetAttendanceByUserIDAndDate(userID uuid.UUID, date time.Time) (*domain.Attendance, error)
	GetAttendancesByUserIDAndPayrollPeriodID(userID uuid.UUID, payrollPeriodID uuid.UUID) ([]*domain.Attendance, error)
	GetAttendancesByUserIDAndPeriod(userID uuid.UUID, startDate, endDate time.Time) ([]domain.Attendance, error)
	UpdateAttendance(attendance *domain.Attendance) error
	UpdateAttendancesTx(tx *gorm.DB, attendances []domain.Attendance) error
}

// AttendanceGormRepository implements repository.AttendanceRepository using GORM.
type AttendanceGormRepository struct {
	db *gorm.DB
}

// NewAttendanceGormRepository creates a new AttendanceGormRepository.
func NewAttendanceGormRepository(db *gorm.DB) AttendanceRepository {
	return &AttendanceGormRepository{db: db}
}

// CreateAttendance creates a new attendance record in the database.
func (r *AttendanceGormRepository) CreateAttendance(attendance *domain.Attendance) error {
	return r.db.Create(attendance).Error
}

// GetAttendanceByID retrieves an attendance record by its ID.
func (r *AttendanceGormRepository) GetAttendanceByID(id uuid.UUID) (*domain.Attendance, error) {
	var attendance domain.Attendance
	err := r.db.First(&attendance, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &attendance, err
}

// GetAttendanceByUserIDAndDate retrieves an attendance record by user ID and date.
func (r *AttendanceGormRepository) GetAttendanceByUserIDAndDate(userID uuid.UUID, date time.Time) (*domain.Attendance, error) {
	var attendance domain.Attendance
	err := r.db.Where("user_id = ? AND date = ?", userID, date.Format("2006-01-02")).First(&attendance).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &attendance, err
}

// GetAttendancesByUserIDAndPeriod retrieves attendance records for a user within a date range.
func (r *AttendanceGormRepository) GetAttendancesByUserIDAndPeriod(userID uuid.UUID, startDate, endDate time.Time) ([]domain.Attendance, error) {
	var attendances []domain.Attendance
	err := r.db.Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).Find(&attendances).Error
	return attendances, err
}

// GetAttendancesByUserIDAndPayrollPeriodID retrieves attendance records for a user within a date range.
func (r *AttendanceGormRepository) GetAttendancesByUserIDAndPayrollPeriodID(userID uuid.UUID, payrollPeriodID uuid.UUID) ([]*domain.Attendance, error) {
	attendances := make([]*domain.Attendance, 0)
	err := r.db.Where("user_id = ? AND payroll_period_id = ?", userID, payrollPeriodID).Find(&attendances).Error
	return attendances, err
}

// UpdateAttendance updates an existing attendance record in the database.
func (r *AttendanceGormRepository) UpdateAttendance(attendance *domain.Attendance) error {
	return r.db.Save(attendance).Error
}

// UpdateAttendancesTx updates multiple attendance records within the given transaction.
func (r *AttendanceGormRepository) UpdateAttendancesTx(tx *gorm.DB, attendances []domain.Attendance) error {
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	for _, attendance := range attendances {
		if err := tx.Save(&attendance).Error; err != nil {
			return err
		}
	}
	return nil
}
