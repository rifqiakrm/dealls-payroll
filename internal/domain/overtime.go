package domain

import (
	"time"

	"github.com/google/uuid"
)

// Overtime records an employee's overtime hours.
type Overtime struct {
	BaseModel
	UserID          uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	User            User           `gorm:"foreignKey:UserID" json:"user"`
	Date            time.Time      `gorm:"type:date;not null" json:"date"`
	Hours           float64        `gorm:"type:numeric;not null" json:"hours"`
	PayrollPeriodID *uuid.UUID     `gorm:"type:uuid" json:"payroll_period_id,omitempty"` // Nullable, set after payroll run
	PayrollPeriod   *PayrollPeriod `gorm:"foreignKey:PayrollPeriodID" json:"payroll_period,omitempty"`
}
