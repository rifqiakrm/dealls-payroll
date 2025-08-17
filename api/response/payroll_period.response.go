package response

import (
	"fmt"
	"time"

	"payroll-system/internal/domain"
)

// PayrollPeriodResponse is the prettified response for payroll period
type PayrollPeriodResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	IsProcessed bool    `json:"is_processed"`
	ProcessedAt *string `json:"processed_at,omitempty"`
}

// ToPayrollPeriodResponse converts domain.PayrollPeriod -> PayrollPeriodResponse
func ToPayrollPeriodResponse(p *domain.PayrollPeriod) PayrollPeriodResponse {
	start := p.StartDate.Format("2 Jan 2006")
	end := p.EndDate.Format("2 Jan 2006")

	var processedAt *string
	if p.ProcessedAt != nil {
		s := p.ProcessedAt.Format(time.RFC3339)
		processedAt = &s
	}

	return PayrollPeriodResponse{
		ID:          p.ID.String(),
		Name:        fmt.Sprintf("Payslip Period %s - %s", start, end),
		StartDate:   p.StartDate.Format("2006-01-02"),
		EndDate:     p.EndDate.Format("2006-01-02"),
		IsProcessed: p.IsProcessed,
		ProcessedAt: processedAt,
	}
}

// ToPayrollPeriodListResponse converts []domain.PayrollPeriod -> []PayrollPeriodResponse
func ToPayrollPeriodListResponse(periods []domain.PayrollPeriod) []PayrollPeriodResponse {
	res := make([]PayrollPeriodResponse, len(periods))
	for i, p := range periods {
		res[i] = ToPayrollPeriodResponse(&p)
	}
	return res
}
