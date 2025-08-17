package domain

// User represents a user in the system, either an employee or an admin.
type User struct {
	BaseModel
	Username string `gorm:"type:varchar(255);uniqueIndex;not null" json:"username"`
	Password string `gorm:"type:varchar(255);not null" json:"-"`   // Stored hashed
	Role     string `gorm:"type:varchar(50);not null" json:"role"` // "employee" or "admin"
}
