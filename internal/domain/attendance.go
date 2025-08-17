package domain

import (
	"time"

	"github.com/google/uuid"
)

// Attendance records an employee's daily attendance.
type Attendance struct {
	BaseModel
	UserID          uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	Date            time.Time      `gorm:"type:date;uniqueIndex:idx_user_date;not null" json:"date"`
	CheckInTime     time.Time      `gorm:"type:time;not null" json:"check_in_time"`
	CheckOutTime    time.Time      `gorm:"type:time;not null" json:"check_out_time"`
	PayrollPeriodID *uuid.UUID     `gorm:"type:uuid" json:"payroll_period_id,omitempty"` // Nullable, set after payroll run
	PayrollPeriod   *PayrollPeriod `gorm:"foreignKey:PayrollPeriodID" json:"payroll_period,omitempty"`
}
