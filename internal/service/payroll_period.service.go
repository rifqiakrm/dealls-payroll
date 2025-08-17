package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"payroll-system/internal/domain"
	"payroll-system/internal/repository"
)

// PayrollPeriodServiceInterface defines the methods of PayrollPeriodService for mocking purposes.
//
//go:generate mockgen -source=payroll_period.service.go -destination=../../tests/mocks/service/mock_payroll_period_service.go -package=mocks
type PayrollPeriodServiceInterface interface {
	// CreatePayrollPeriod creates a new payroll period.
	CreatePayrollPeriod(startDate, endDate time.Time, createdBy uuid.UUID, ipAddress, requestID string) (*domain.PayrollPeriod, error)
	// GetPayrollPeriodByID retrieves a payroll period by its ID.
	GetPayrollPeriodByID(id uuid.UUID) (*domain.PayrollPeriod, error)
	// GetAllPayrollPeriods retrieves all payroll periods.
	GetAllPayrollPeriods() ([]domain.PayrollPeriod, error)
	// MarkPayrollPeriodAsProcessed marks a payroll period as processed.
	MarkPayrollPeriodAsProcessed(id uuid.UUID, updatedBy uuid.UUID, ipAddress string) error
}

// PayrollPeriodService provides business logic for payroll period management.
type PayrollPeriodService struct {
	payrollPeriodRepo repository.PayrollPeriodRepository
	auditRepo         repository.AuditLogRepository
}

// NewPayrollPeriodService creates a new PayrollPeriodService.
func NewPayrollPeriodService(
	payrollPeriodRepo repository.PayrollPeriodRepository,
	auditRepo repository.AuditLogRepository,
) *PayrollPeriodService {
	return &PayrollPeriodService{
		payrollPeriodRepo: payrollPeriodRepo,
		auditRepo:         auditRepo,
	}
}

// CreatePayrollPeriod creates a new payroll period.
func (s *PayrollPeriodService) CreatePayrollPeriod(
	startDate, endDate time.Time,
	createdBy uuid.UUID,
	ipAddress string,
	requestID string,
) (*domain.PayrollPeriod, error) {
	// Validation
	if !endDate.After(startDate) {
		return nil, errors.New("end date must be after start date")
	}

	// Check for overlaps
	overlappingPeriods, err := s.payrollPeriodRepo.GetOverlappingPayrollPeriods(startDate, endDate)
	if err != nil {
		return nil, err
	}
	if len(overlappingPeriods) > 0 {
		return nil, errors.New("a payroll period overlapping with these dates already exists")
	}

	period := &domain.PayrollPeriod{
		StartDate:   startDate,
		EndDate:     endDate,
		IsProcessed: false,
		BaseModel: domain.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: createdBy,
			UpdatedBy: createdBy,
			IPAddress: ipAddress,
		},
	}

	if err := s.payrollPeriodRepo.CreatePayrollPeriod(period); err != nil {
		return nil, err
	}

	// Audit log
	_ = repository.CreateAuditLog(
		s.auditRepo,
		&createdBy,
		"CREATE",
		"PayrollPeriod",
		&period.ID,
		nil,
		period,
		ipAddress,
		requestID,
	)

	return period, nil
}

// GetPayrollPeriodByID retrieves a payroll period by its ID.
func (s *PayrollPeriodService) GetPayrollPeriodByID(id uuid.UUID) (*domain.PayrollPeriod, error) {
	return s.payrollPeriodRepo.GetPayrollPeriodByID(id)
}

// GetAllPayrollPeriods retrieves all payroll periods.
func (s *PayrollPeriodService) GetAllPayrollPeriods() ([]domain.PayrollPeriod, error) {
	return s.payrollPeriodRepo.GetAllPayrollPeriods()
}

// MarkPayrollPeriodAsProcessed marks a payroll period as processed.
func (s *PayrollPeriodService) MarkPayrollPeriodAsProcessed(id uuid.UUID, updatedBy uuid.UUID, ipAddress string) error {
	period, err := s.payrollPeriodRepo.GetPayrollPeriodByID(id)
	if err != nil {
		return err
	}
	if period == nil {
		return errors.New("payroll period not found")
	}

	// Update the period
	if err := s.payrollPeriodRepo.MarkPayrollPeriodAsProcessed(id); err != nil {
		return err
	}

	// Audit log
	requestID := uuid.New().String()
	_ = repository.CreateAuditLog(
		s.auditRepo,
		&updatedBy,
		"UPDATE",
		"PayrollPeriod",
		&period.ID,
		period,
		period,
		ipAddress,
		requestID,
	)

	return nil
}
