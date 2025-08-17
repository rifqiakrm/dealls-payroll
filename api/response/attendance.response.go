package response

import (
	"payroll-system/internal/domain"
)

// AttendanceResponse defines how attendance data is returned to the client.
type AttendanceResponse struct {
	ID              string  `json:"id"`
	Date            string  `json:"date"`           // formatted YYYY-MM-DD
	CheckInTime     string  `json:"check_in_time"`  // formatted HH:MM:SS
	CheckOutTime    string  `json:"check_out_time"` // formatted HH:MM:SS
	HoursWorked     float64 `json:"hours_worked"`
	PayrollPeriodID *string `json:"payroll_period_id,omitempty"`
}

// ToAttendanceResponse maps domain.Attendance -> AttendanceResponse
func ToAttendanceResponse(a *domain.Attendance) AttendanceResponse {
	var payrollPeriodID *string
	if a.PayrollPeriodID != nil {
		id := a.PayrollPeriodID.String()
		payrollPeriodID = &id
	}

	hours := a.CheckOutTime.Sub(a.CheckInTime).Hours()

	if hours > 8 {
		hours = 8
	} else if hours < 0 {
		hours = 0
	}

	return AttendanceResponse{
		ID:              a.ID.String(),
		Date:            a.Date.Format("2006-01-02"),
		CheckInTime:     a.CheckInTime.Format("2006-01-02 15:04:05"),
		CheckOutTime:    a.CheckOutTime.Format("2006-01-02 15:04:05"),
		HoursWorked:     hours,
		PayrollPeriodID: payrollPeriodID,
	}
}
