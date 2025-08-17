package response

import (
	"payroll-system/internal/domain"
)

// ReimbursementResponse defines the structure returned to the client.
type ReimbursementResponse struct {
	ID              string  `json:"id"`
	UserID          string  `json:"user_id"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	PayrollPeriodID *string `json:"payroll_period_id,omitempty"`
}

// ToReimbursementResponse maps domain.Reimbursement -> ReimbursementResponse
func ToReimbursementResponse(r *domain.Reimbursement) ReimbursementResponse {
	var periodID *string
	if r.PayrollPeriodID != nil {
		id := r.PayrollPeriodID.String()
		periodID = &id
	}

	return ReimbursementResponse{
		ID:              r.ID.String(),
		UserID:          r.UserID.String(),
		Amount:          r.Amount,
		Description:     r.Description,
		PayrollPeriodID: periodID,
	}
}
