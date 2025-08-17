package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"payroll-system/internal/domain"
	"payroll-system/internal/repository"
)

const MaxOvertimeHoursPerDay = 3.0

// OvertimeServiceInterface defines the methods of OvertimeService for mocking purposes.
//
//go:generate mockgen -source=overtime.service.go -destination=../../tests/mocks/service/mock_overtime_service.go -package=mocks
type OvertimeServiceInterface interface {
	// SubmitOvertime allows an employee to submit their overtime hours.
	SubmitOvertime(userID uuid.UUID, date time.Time, hours float64, ipAddress, requestID string) (*domain.Overtime, error)
}

// OvertimeService provides business logic for overtime management.
type OvertimeService struct {
	overtimeRepo repository.OvertimeRepository
	auditRepo    repository.AuditLogRepository
}

// NewOvertimeService creates a new OvertimeService.
func NewOvertimeService(overtimeRepo repository.OvertimeRepository, auditRepo repository.AuditLogRepository) *OvertimeService {
	return &OvertimeService{
		overtimeRepo: overtimeRepo,
		auditRepo:    auditRepo,
	}
}

// SubmitOvertime allows an employee to submit their overtime hours.
func (s *OvertimeService) SubmitOvertime(userID uuid.UUID, date time.Time, hours float64, ipAddress string, requestID string) (*domain.Overtime, error) {
	// Rule: Overtime cannot be more than MaxOvertimeHoursPerDay per day.
	existingOvertimes, err := s.overtimeRepo.GetOvertimeByUserIDAndDate(userID, date)
	if err != nil {
		return nil, err
	}

	totalHoursToday := 0.0
	for _, ot := range existingOvertimes {
		totalHoursToday += ot.Hours
	}

	if totalHoursToday+hours > MaxOvertimeHoursPerDay {
		return nil, fmt.Errorf("total overtime hours for %s cannot exceed %.1f hours", date.Format("2006-01-02"), MaxOvertimeHoursPerDay)
	}

	newOvertime := &domain.Overtime{
		UserID: userID,
		Date:   date,
		Hours:  hours,
		BaseModel: domain.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: userID,
			UpdatedBy: userID,
			IPAddress: ipAddress,
		},
	}

	o, err := s.overtimeRepo.CreateOvertime(newOvertime)

	if err != nil {
		return nil, err
	}

	// Audit log for overtime submission
	_ = repository.CreateAuditLog(
		s.auditRepo,
		&userID,
		"CREATE",
		"Overtime",
		&o.ID,
		nil,
		newOvertime,
		ipAddress,
		requestID,
	)

	return newOvertime, nil
}
