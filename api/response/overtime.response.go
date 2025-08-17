package response

import (
	"payroll-system/internal/domain"
)

// OvertimeResponse defines how overtime data is returned to the client.
type OvertimeResponse struct {
	ID              string  `json:"id"`
	Date            string  `json:"date"` // formatted YYYY-MM-DD
	Hours           float64 `json:"hours"`
	PayrollPeriodID *string `json:"payroll_period_id,omitempty"`
}

// ToOvertimeResponse maps domain.Overtime -> OvertimeResponse
func ToOvertimeResponse(o *domain.Overtime) OvertimeResponse {
	var payrollPeriodID *string
	if o.PayrollPeriodID != nil {
		id := o.PayrollPeriodID.String()
		payrollPeriodID = &id
	}

	return OvertimeResponse{
		ID:              o.ID.String(),
		Date:            o.Date.Format("2006-01-02"),
		Hours:           o.Hours,
		PayrollPeriodID: payrollPeriodID,
	}
}
