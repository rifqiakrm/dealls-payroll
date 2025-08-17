package domain

import (
	"github.com/google/uuid"
)

// Reimbursement records an employee's reimbursement request.
type Reimbursement struct {
	BaseModel
	UserID        uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	User          User           `gorm:"foreignKey:UserID" json:"user"`
	Amount        float64        `gorm:"type:numeric;not null" json:"amount"`
	Description   string         `gorm:"type:text" json:"description"`
	PayrollPeriodID *uuid.UUID     `gorm:"type:uuid" json:"payroll_period_id,omitempty"` // Nullable, set after payroll run
	PayrollPeriod   *PayrollPeriod `gorm:"foreignKey:PayrollPeriodID" json:"payroll_period,omitempty"`
}
