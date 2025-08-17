package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"payroll-system/internal/domain"
	"payroll-system/internal/repository"
)

// AttendanceServiceInterface defines the methods of AttendanceService for mocking purposes.
//
//go:generate mockgen -source=attendance.service.go -destination=../../tests/mocks/service/mock_attendance_service.go -package=mocks
type AttendanceServiceInterface interface {
	// SubmitAttendance allows an employee to submit their attendance.
	SubmitAttendance(userID uuid.UUID, checkInTime, checkOutTime time.Time, ipAddress string, requestID string) (*domain.Attendance, error)
}

// AttendanceService provides business logic for attendance management.
type AttendanceService struct {
	attendanceRepo repository.AttendanceRepository
	auditRepo      repository.AuditLogRepository
}

// NewAttendanceService creates a new AttendanceService.
func NewAttendanceService(attendanceRepo repository.AttendanceRepository, auditRepo repository.AuditLogRepository) *AttendanceService {
	return &AttendanceService{
		attendanceRepo: attendanceRepo,
		auditRepo:      auditRepo,
	}
}

// SubmitAttendance allows an employee to submit their attendance.
// It handles both check-in and check-out, and updates existing records for the same day.
func (s *AttendanceService) SubmitAttendance(userID uuid.UUID, checkInTime, checkOutTime time.Time, ipAddress string, requestID string) (*domain.Attendance, error) {
	// Rule: Users cannot submit on weekends.
	if checkInTime.Weekday() == time.Saturday || checkInTime.Weekday() == time.Sunday {
		return nil, errors.New("attendance cannot be submitted on weekends")
	}

	now := time.Now()

	// Check if an attendance record already exists for this user and date.
	existingAttendance, err := s.attendanceRepo.GetAttendanceByUserIDAndDate(userID, checkInTime)
	if err != nil {
		return nil, err
	}

	if existingAttendance != nil {
		// Update existing record
		oldValue := *existingAttendance
		existingAttendance.CheckInTime = checkInTime
		existingAttendance.CheckOutTime = checkOutTime
		existingAttendance.UpdatedAt = now
		existingAttendance.UpdatedBy = userID
		existingAttendance.IPAddress = ipAddress

		if err := s.attendanceRepo.UpdateAttendance(existingAttendance); err != nil {
			return nil, err
		}

		// Create audit log
		_ = repository.CreateAuditLog(s.auditRepo, &userID, "UPDATE", "Attendance", &existingAttendance.ID, oldValue, existingAttendance, ipAddress, requestID)
		return existingAttendance, nil
	}

	// Create new attendance record
	newAttendance := &domain.Attendance{
		UserID:       userID,
		Date:         checkInTime,
		CheckInTime:  checkInTime,
		CheckOutTime: checkOutTime,
		BaseModel: domain.BaseModel{
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: userID,
			UpdatedBy: userID,
			IPAddress: ipAddress,
		},
	}

	if err := s.attendanceRepo.CreateAttendance(newAttendance); err != nil {
		return nil, err
	}

	// Create audit log for creation
	_ = repository.CreateAuditLog(s.auditRepo, &userID, "CREATE", "Attendance", &newAttendance.ID, nil, newAttendance, ipAddress, requestID)
	return newAttendance, nil
}
