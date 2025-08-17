package domain

import (
	"time"
)

// PayrollPeriod defines the start and end dates for a payroll cycle.
type PayrollPeriod struct {
	BaseModel
	StartDate   time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate     time.Time `gorm:"type:date;not null" json:"end_date"`
	IsProcessed bool      `gorm:"default:false;not null" json:"is_processed"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"` // Nullable
}
