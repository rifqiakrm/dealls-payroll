package response

import (
	"payroll-system/internal/domain"
	"time"
)

const RegularWorkingHoursPerDay = 8
const OvertimeMultiplier = 2.0

// AttendancePayslipResponse defines how attendance data is returned to the client.
type AttendancePayslipResponse struct {
	ID              string  `json:"id"`
	Date            string  `json:"date"`           // formatted YYYY-MM-DD
	CheckInTime     string  `json:"check_in_time"`  // formatted HH:MM:SS
	CheckOutTime    string  `json:"check_out_time"` // formatted HH:MM:SS
	HoursWorked     float64 `json:"hours_worked"`
	BasePay         float64 `json:"base_pay"`
	PayrollPeriodID *string `json:"payroll_period_id,omitempty"`
}

// OvertimePayslipResponse defines how overtime data is returned to the client.
type OvertimePayslipResponse struct {
	ID              string  `json:"id"`
	Date            string  `json:"date"` // formatted YYYY-MM-DD
	Hours           float64 `json:"hours"`
	BasePay         float64 `json:"base_pay"`
	PayrollPeriodID *string `json:"payroll_period_id,omitempty"`
}

// PayslipResponse defines the structure returned to the client.
type PayslipResponse struct {
	ID                 string      `json:"id"`
	UserID             string      `json:"user_id"`
	PayrollPeriodID    string      `json:"payroll_period_id"`
	BaseSalary         float64     `json:"base_salary"`
	ProratedSalary     float64     `json:"prorated_salary"`
	OvertimePay        float64     `json:"overtime_pay"`
	TotalReimbursement float64     `json:"total_reimbursement"`
	TotalTakeHomePay   float64     `json:"total_take_home_pay"`
	Overtimes          interface{} `json:"overtimes"`
	Attendances        interface{} `json:"attendances"`
}

// ToPayslipResponse maps domain.Payslip -> PayslipResponse
func ToPayslipResponse(p *domain.Payslip) PayslipResponse {
	// Calculate hourly pay
	totalPossibleWorkingHours := 0.0
	for d := p.PayrollPeriod.StartDate; !d.After(p.PayrollPeriod.EndDate); d = d.Add(24 * time.Hour) {
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			totalPossibleWorkingHours += RegularWorkingHoursPerDay
		}
	}

	var hourlyRate float64

	if totalPossibleWorkingHours > 0 {
		hourlyRate = p.BaseSalary / totalPossibleWorkingHours
	}

	overtimes := make([]OvertimePayslipResponse, 0)

	for _, o := range p.Overtimes {
		id := o.PayrollPeriodID.String()
		payrollPeriodID := &id

		basePay := o.Hours * hourlyRate * OvertimeMultiplier

		overtimes = append(overtimes, OvertimePayslipResponse{
			ID:              o.ID.String(),
			Date:            o.Date.Format("2006-01-02"),
			Hours:           o.Hours,
			BasePay:         basePay,
			PayrollPeriodID: payrollPeriodID,
		})
	}

	attendances := make([]AttendancePayslipResponse, 0)

	for _, a := range p.Attendances {
		hours := a.CheckOutTime.Sub(a.CheckInTime).Hours()

		if hours > 8 {
			hours = 8
		} else if hours < 0 {
			hours = 0
		}

		id := a.PayrollPeriodID.String()
		payrollPeriodID := &id

		attendances = append(attendances, AttendancePayslipResponse{
			ID:              a.ID.String(),
			Date:            a.Date.Format("2006-01-02"),
			CheckInTime:     a.CheckInTime.Format("2006-01-02 15:04:05"),
			CheckOutTime:    a.CheckOutTime.Format("2006-01-02 15:04:05"),
			HoursWorked:     hours,
			BasePay:         hourlyRate,
			PayrollPeriodID: payrollPeriodID,
		})
	}

	return PayslipResponse{
		ID:                 p.ID.String(),
		UserID:             p.UserID.String(),
		PayrollPeriodID:    p.PayrollPeriodID.String(),
		BaseSalary:         p.BaseSalary,
		ProratedSalary:     p.ProratedSalary,
		OvertimePay:        p.OvertimePay,
		TotalReimbursement: p.TotalReimbursement,
		TotalTakeHomePay:   p.TotalTakeHomePay,
		Overtimes:          overtimes,
		Attendances:        attendances,
	}
}
