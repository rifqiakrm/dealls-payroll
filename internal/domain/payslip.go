package domain

import (
	"github.com/google/uuid"
)

// Payslip stores the calculated payslip details for an employee.
type Payslip struct {
	BaseModel
	UserID             uuid.UUID     `gorm:"type:uuid;not null" json:"user_id"`
	User               User          `gorm:"foreignKey:UserID" json:"user"`
	PayrollPeriodID    uuid.UUID     `gorm:"type:uuid;not null" json:"payroll_period_id"`
	PayrollPeriod      PayrollPeriod `gorm:"foreignKey:PayrollPeriodID" json:"payroll_period"`
	Overtimes          []*Overtime   `gorm:"-" json:"overtimes"`
	Attendances        []*Attendance `gorm:"-" json:"attendances"`
	BaseSalary         float64       `gorm:"type:numeric;not null" json:"base_salary"`
	ProratedSalary     float64       `gorm:"type:numeric;not null" json:"prorated_salary"`
	OvertimePay        float64       `gorm:"type:numeric;not null" json:"overtime_pay"`
	TotalReimbursement float64       `gorm:"type:numeric;not null" json:"total_reimbursement"`
	TotalTakeHomePay   float64       `gorm:"type:numeric;not null" json:"total_take_home_pay"`
}
