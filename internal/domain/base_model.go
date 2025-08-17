package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all entities.
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	CreatedBy uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	UpdatedBy uuid.UUID      `gorm:"type:uuid" json:"updated_by"`
	IPAddress string         `gorm:"type:varchar(45)" json:"ip_address"` // IPv4 or IPv6
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}
