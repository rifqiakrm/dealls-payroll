package service

import (
	"errors"

	"github.com/google/uuid"

	"payroll-system/internal/domain"
	"payroll-system/internal/repository"
)

// PayslipServiceInterface defines the methods of PayslipService for mocking purposes.
//
//go:generate mockgen -source=payslip.service.go -destination=../../tests/mocks/service/mock_payslip_service.go -package=mocks
type PayslipServiceInterface interface {
	// GetEmployeePayslip retrieves a payslip for a specific employee and payroll period.
	GetEmployeePayslip(userID, periodID uuid.UUID) (*domain.Payslip, error)
	// GetPayslipSummaryForPeriod retrieves a summary of all payslips for a given payroll period.
	GetPayslipSummaryForPeriod(periodID uuid.UUID) ([]domain.Payslip, float64, error)
}

// PayslipService provides business logic for payslip generation.
type PayslipService struct {
	payslipRepo       repository.PayslipRepository
	payslipPeriodRepo repository.PayrollPeriodRepository
	attendanceRepo    repository.AttendanceRepository
	overtimeRepo      repository.OvertimeRepository
}

// NewPayslipService creates a new PayslipService.
func NewPayslipService(
	payslipRepo repository.PayslipRepository,
	payslipPeriodRepo repository.PayrollPeriodRepository,
	attendanceRepo repository.AttendanceRepository,
	overtimeRepo repository.OvertimeRepository,
) *PayslipService {
	return &PayslipService{
		payslipRepo:       payslipRepo,
		payslipPeriodRepo: payslipPeriodRepo,
		attendanceRepo:    attendanceRepo,
		overtimeRepo:      overtimeRepo,
	}
}

// GetEmployeePayslip retrieves a payslip for a specific employee and payroll period.
func (s *PayslipService) GetEmployeePayslip(userID, periodID uuid.UUID) (*domain.Payslip, error) {
	period, err := s.payslipPeriodRepo.GetPayrollPeriodByID(periodID)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, errors.New("payroll period not found")
	}
	if !period.IsProcessed {
		return nil, errors.New("payslip can only be generated for processed payroll periods")
	}

	payslip, err := s.payslipRepo.GetPayslipByUserIDAndPeriodID(userID, periodID)
	if err != nil {
		return nil, err
	}
	if payslip == nil {
		return nil, errors.New("payslip not found for this user and period")
	}

	// Attach related data
	payslip.PayrollPeriod = *period

	attendances, err := s.attendanceRepo.GetAttendancesByUserIDAndPayrollPeriodID(userID, periodID)
	if err != nil {
		return nil, err
	}
	payslip.Attendances = attendances

	overtimes, err := s.overtimeRepo.GetOvertimesByUserIDAndPayrollPeriodID(userID, periodID)
	if err != nil {
		return nil, err
	}
	payslip.Overtimes = overtimes

	return payslip, nil
}

// GetPayslipSummaryForPeriod retrieves a summary of all payslips for a given payroll period.
func (s *PayslipService) GetPayslipSummaryForPeriod(periodID uuid.UUID) ([]domain.Payslip, float64, error) {
	period, err := s.payslipPeriodRepo.GetPayrollPeriodByID(periodID)
	if err != nil {
		return nil, 0, err
	}
	if period == nil {
		return nil, 0, errors.New("payroll period not found")
	}
	if !period.IsProcessed {
		return nil, 0, errors.New("payslip summary can only be generated for processed payroll periods")
	}

	payslips, err := s.payslipRepo.GetAllPayslipsByPeriodID(periodID)
	if err != nil {
		return nil, 0, err
	}

	var totalTakeHomePay float64
	resultPayslips := make([]domain.Payslip, 0, len(payslips))

	for _, p := range payslips {
		p.PayrollPeriod = *period

		attendances, err := s.attendanceRepo.GetAttendancesByUserIDAndPayrollPeriodID(p.UserID, periodID)
		if err != nil {
			return nil, 0, err
		}
		p.Attendances = attendances

		overtimes, err := s.overtimeRepo.GetOvertimesByUserIDAndPayrollPeriodID(p.UserID, periodID)
		if err != nil {
			return nil, 0, err
		}
		p.Overtimes = overtimes

		totalTakeHomePay += p.TotalTakeHomePay
		resultPayslips = append(resultPayslips, p)
	}

	return resultPayslips, totalTakeHomePay, nil
}
