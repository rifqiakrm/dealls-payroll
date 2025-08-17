package domain

import (
	"github.com/google/uuid"
)

// EmployeeProfile stores additional details for an employee.
type EmployeeProfile struct {
	BaseModel
	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID" json:"user"`
	Salary float64   `gorm:"type:numeric;not null" json:"salary"`
}
