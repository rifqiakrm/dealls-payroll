package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AuditLog tracks significant changes and actions in the system.
type AuditLog struct {
	BaseModel
	UserID     *uuid.UUID     `gorm:"type:uuid" json:"user_id,omitempty"` // Nullable if system action
	User       *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action     string         `gorm:"type:varchar(255);not null" json:"action"`      // e.g., "CREATE", "UPDATE", "DELETE", "LOGIN"
	EntityName string         `gorm:"type:varchar(255);not null" json:"entity_name"` // e.g., "User", "Attendance"
	EntityID   *uuid.UUID     `gorm:"type:uuid" json:"entity_id,omitempty"`          // ID of the affected entity
	OldValue   datatypes.JSON `gorm:"type:jsonb" json:"old_value,omitempty"`         // JSON representation of old state
	NewValue   datatypes.JSON `gorm:"type:jsonb" json:"new_value,omitempty"`         // JSON representation of new state
	RequestID  string         `gorm:"type:varchar(255);not null" json:"request_id"`
	Timestamp  time.Time      `gorm:"not null" json:"timestamp"`
}
