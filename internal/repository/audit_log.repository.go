package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// AuditLogRepository defines the interface for audit log operations.
//
//go:generate mockgen -source=audit_log.repository.go -destination=../../tests/mocks/repository/mock_audit_log_repository.go -package=mocks
type AuditLogRepository interface {
	Create(audit *domain.AuditLog) error
	GetByID(id uuid.UUID) (*domain.AuditLog, error)
	GetAllByUser(userID uuid.UUID, limit int) ([]domain.AuditLog, error)
}

// AuditLogGormRepository implements repository.AuditLogRepository using GORM.
type AuditLogGormRepository struct {
	db *gorm.DB
}

// NewAuditLogGormRepository creates a new AuditLogGormRepository.
func NewAuditLogGormRepository(db *gorm.DB) AuditLogRepository {
	return &AuditLogGormRepository{db: db}
}

// Create inserts a new audit log record.
func (r *AuditLogGormRepository) Create(audit *domain.AuditLog) error {
	if audit.Timestamp.IsZero() {
		audit.Timestamp = time.Now()
	}
	return r.db.Create(audit).Error
}

// GetByID retrieves an audit log record by its ID.
func (r *AuditLogGormRepository) GetByID(id uuid.UUID) (*domain.AuditLog, error) {
	var audit domain.AuditLog
	err := r.db.First(&audit, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &audit, err
}

// GetAllByUser retrieves audit logs for a specific user, limited by 'limit'.
func (r *AuditLogGormRepository) GetAllByUser(userID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	query := r.db.Where("user_id = ?", userID).Order("timestamp desc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}
