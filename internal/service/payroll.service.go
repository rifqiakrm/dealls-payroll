package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
	"payroll-system/internal/repository"
)

const (
	RegularWorkingHoursPerDay = 8
	WorkingDaysPerWeek        = 5
	WorkingDaysPerMonth       = 20 // Approximation for monthly-based pay
	OvertimeMultiplier        = 2.0
)

// PayrollServiceInterface defines methods of PayrollService for mocking purposes.
//
//go:generate mockgen -source=payroll.service.go -destination=../../tests/mocks/service/mock_payroll_service.go -package=mocks
type PayrollServiceInterface interface {
	// RunPayroll processes payroll for a given payroll period.
	RunPayroll(periodID uuid.UUID, processedBy uuid.UUID, ipAddress, requestID string) error
	// CalculatePayslip calculates payslip and related records for a user.
	CalculatePayslip(userID uuid.UUID, period *domain.PayrollPeriod, processedBy uuid.UUID, ipAddress string) (*domain.Payslip, []domain.Attendance, []domain.Overtime, []domain.Reimbursement, error)
}

// PayrollService provides business logic for payroll processing.
type PayrollService struct {
	payslipRepo         repository.PayslipRepository
	payrollPeriodRepo   repository.PayrollPeriodRepository
	employeeProfileRepo repository.EmployeeProfileRepository
	attendanceRepo      repository.AttendanceRepository
	overtimeRepo        repository.OvertimeRepository
	reimbursementRepo   repository.ReimbursementRepository
	auditRepo           repository.AuditLogRepository
	db                  *gorm.DB // For transaction management
}

// NewPayrollService creates a new PayrollService.
func NewPayrollService(
	payslipRepo repository.PayslipRepository,
	payrollPeriodRepo repository.PayrollPeriodRepository,
	employeeProfileRepo repository.EmployeeProfileRepository,
	attendanceRepo repository.AttendanceRepository,
	overtimeRepo repository.OvertimeRepository,
	reimbursementRepo repository.ReimbursementRepository,
	auditRepo repository.AuditLogRepository,
	db *gorm.DB,
) *PayrollService {
	return &PayrollService{
		payslipRepo:         payslipRepo,
		payrollPeriodRepo:   payrollPeriodRepo,
		employeeProfileRepo: employeeProfileRepo,
		attendanceRepo:      attendanceRepo,
		overtimeRepo:        overtimeRepo,
		reimbursementRepo:   reimbursementRepo,
		auditRepo:           auditRepo,
		db:                  db,
	}
}

func (s *PayrollService) RunPayroll(periodID uuid.UUID, processedBy uuid.UUID, ipAddress string, requestID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		period, err := s.payrollPeriodRepo.GetPayrollPeriodByID(periodID)
		if err != nil {
			return err
		}
		if period == nil {
			return errors.New("payroll period not found")
		}
		if period.IsProcessed {
			return errors.New("payroll already processed")
		}

		employees, err := s.employeeProfileRepo.GetAllEmployeeProfiles()
		if err != nil {
			return err
		}

		for _, emp := range employees {
			payslip, attendances, overtimes, reimbursements, err := s.CalculatePayslip(emp.UserID, period, processedBy, ipAddress)
			if err != nil {
				return fmt.Errorf("failed to calculate payslip for user %s: %w", emp.UserID, err)
			}

			// Save payslip
			if err := s.payslipRepo.CreatePayslipTx(tx, payslip); err != nil {
				return fmt.Errorf("failed to save payslip for user %s: %w", emp.UserID, err)
			}

			// Update related records
			if err := s.attendanceRepo.UpdateAttendancesTx(tx, attendances); err != nil {
				return fmt.Errorf("failed to update attendances for user %s: %w", emp.UserID, err)
			}
			if err := s.overtimeRepo.UpdateOvertimesTx(tx, overtimes); err != nil {
				return fmt.Errorf("failed to update overtimes for user %s: %w", emp.UserID, err)
			}
			if err := s.reimbursementRepo.UpdateReimbursementsTx(tx, reimbursements); err != nil {
				return fmt.Errorf("failed to update reimbursements for user %s: %w", emp.UserID, err)
			}

			// Audit log for payslip creation
			_ = repository.CreateAuditLog(
				s.auditRepo,
				&processedBy,
				"CREATE",
				"Payslip",
				&payslip.ID,
				nil,
				payslip,
				ipAddress,
				requestID,
			)
		}

		// Mark payroll as processed
		if err := s.payrollPeriodRepo.MarkPayrollPeriodAsProcessedTx(tx, periodID); err != nil {
			return fmt.Errorf("failed to mark payroll period as processed: %w", err)
		}

		// Audit log for payroll period processing
		_ = repository.CreateAuditLog(
			s.auditRepo,
			&processedBy,
			"UPDATE",
			"PayrollPeriod",
			&period.ID,
			nil,
			period,
			ipAddress,
			requestID,
		)

		return nil
	})
}

func (s *PayrollService) CalculatePayslip(
	userID uuid.UUID,
	period *domain.PayrollPeriod,
	processedBy uuid.UUID,
	ipAddress string,
) (*domain.Payslip, []domain.Attendance, []domain.Overtime, []domain.Reimbursement, error) {

	empProfile, err := s.employeeProfileRepo.GetEmployeeProfileByUserID(userID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if empProfile == nil {
		return nil, nil, nil, nil, errors.New("employee profile not found")
	}

	baseSalary := empProfile.Salary

	// Attendance
	attendances, err := s.attendanceRepo.GetAttendancesByUserIDAndPeriod(userID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	totalWorkedHours := 0.0
	for _, att := range attendances {
		if (att.Date.After(period.StartDate) || att.Date.Equal(period.StartDate)) &&
			(att.Date.Before(period.EndDate) || att.Date.Equal(period.EndDate)) {

			workedHours := att.CheckOutTime.Sub(att.CheckInTime).Hours()

			if workedHours > RegularWorkingHoursPerDay {
				workedHours = RegularWorkingHoursPerDay // cap at 8h
			} else if workedHours < RegularWorkingHoursPerDay {
				workedHours = 0
			}
			totalWorkedHours += workedHours
		}
	}

	// totalPossibleWorkingHours = working days * 8 hours
	totalPossibleWorkingHours := 0.0
	for d := period.StartDate; !d.After(period.EndDate); d = d.Add(24 * time.Hour) {
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			totalPossibleWorkingHours += RegularWorkingHoursPerDay
		}
	}

	var hourlyRate float64

	proratedSalary := 0.0
	if totalPossibleWorkingHours > 0 {
		hourlyRate = baseSalary / totalPossibleWorkingHours
		proratedSalary = hourlyRate * totalWorkedHours
	}

	// Overtime
	overtimes, err := s.overtimeRepo.GetOvertimesByUserIDAndPeriod(userID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	totalOvertimeHours := 0.0
	for _, ot := range overtimes {
		totalOvertimeHours += ot.Hours
	}

	overtimePay := totalOvertimeHours * hourlyRate * OvertimeMultiplier

	// Reimbursements
	reimbursements, err := s.reimbursementRepo.GetReimbursementsByUserIDAndPeriod(userID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	totalReimbursement := 0.0
	for _, reimb := range reimbursements {
		totalReimbursement += reimb.Amount
	}

	totalTakeHomePay := proratedSalary + overtimePay + totalReimbursement

	payslip := &domain.Payslip{
		UserID:             userID,
		PayrollPeriodID:    period.ID,
		BaseSalary:         baseSalary,
		ProratedSalary:     proratedSalary,
		OvertimePay:        overtimePay,
		TotalReimbursement: totalReimbursement,
		TotalTakeHomePay:   totalTakeHomePay,
		BaseModel: domain.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: processedBy,
			UpdatedBy: processedBy,
			IPAddress: ipAddress,
		},
	}

	// Attach payroll period to related records (immutability)
	for i := range attendances {
		attendances[i].PayrollPeriodID = &period.ID
		attendances[i].UpdatedAt = time.Now()
		attendances[i].UpdatedBy = processedBy
		attendances[i].IPAddress = ipAddress
	}
	for i := range overtimes {
		overtimes[i].PayrollPeriodID = &period.ID
		overtimes[i].UpdatedAt = time.Now()
		overtimes[i].UpdatedBy = processedBy
		overtimes[i].IPAddress = ipAddress
	}
	for i := range reimbursements {
		reimbursements[i].PayrollPeriodID = &period.ID
		reimbursements[i].UpdatedAt = time.Now()
		reimbursements[i].UpdatedBy = processedBy
		reimbursements[i].IPAddress = ipAddress
	}

	return payslip, attendances, overtimes, reimbursements, nil
}
