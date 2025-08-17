package repository

import (
	"encoding/json"
	"time"

	"payroll-system/internal/domain"

	"github.com/google/uuid"
)

var CreateAuditLogFunc = CreateAuditLog

// CreateAuditLog is a helper to easily insert an audit log.
// - userID can be nil for system actions
// - oldValue and newValue can be any struct, will be marshaled to JSON
func CreateAuditLog(repo AuditLogRepository, userID *uuid.UUID, action, entityName string, entityID *uuid.UUID, oldValue, newValue any, ipAddress string, requestID string) error {
	oldJSON, err := json.Marshal(oldValue)
	if err != nil {
		return err
	}
	newJSON, err := json.Marshal(newValue)
	if err != nil {
		return err
	}

	audit := &domain.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityName: entityName,
		EntityID:   entityID,
		OldValue:   oldJSON,
		NewValue:   newJSON,
		RequestID:  requestID,
		Timestamp:  time.Now(),
		BaseModel: domain.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: *userID,
			UpdatedBy: *userID,
			IPAddress: ipAddress,
		},
	}

	return repo.Create(audit)
}
