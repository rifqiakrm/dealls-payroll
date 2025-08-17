package service

import (
	"time"

	"github.com/google/uuid"

	"payroll-system/internal/domain"
	"payroll-system/internal/repository"
)

// ReimbursementServiceInterface defines the methods of ReimbursementService for mocking purposes.
//
//go:generate mockgen -source=reimbursement.service.go -destination=../../tests/mocks/service/mock_reimbursement_service.go -package=mocks
type ReimbursementServiceInterface interface {
	// SubmitReimbursement allows an employee to submit a reimbursement request.
	SubmitReimbursement(userID uuid.UUID, amount float64, description, ipAddress, requestID string) (*domain.Reimbursement, error)
}

// ReimbursementService provides business logic for reimbursement management.
type ReimbursementService struct {
	reimbursementRepo repository.ReimbursementRepository
	auditLogRepo      repository.AuditLogRepository
}

// NewReimbursementService creates a new ReimbursementService.
func NewReimbursementService(
	reimbursementRepo repository.ReimbursementRepository,
	auditLogRepo repository.AuditLogRepository,
) *ReimbursementService {
	return &ReimbursementService{
		reimbursementRepo: reimbursementRepo,
		auditLogRepo:      auditLogRepo,
	}
}

// SubmitReimbursement allows an employee to submit a reimbursement request.
func (s *ReimbursementService) SubmitReimbursement(
	userID uuid.UUID,
	amount float64,
	description, ipAddress, requestID string,
) (*domain.Reimbursement, error) {

	newReimbursement := &domain.Reimbursement{
		UserID:      userID,
		Amount:      amount,
		Description: description,
		BaseModel: domain.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: userID,
			UpdatedBy: userID,
			IPAddress: ipAddress,
		},
	}

	// Save reimbursement
	if err := s.reimbursementRepo.CreateReimbursement(newReimbursement); err != nil {
		return nil, err
	}

	// Create audit log
	_ = repository.CreateAuditLog(
		s.auditLogRepo,
		&userID,
		"CREATE",
		"Reimbursement",
		&newReimbursement.ID,
		nil, // oldValue is nil for creation
		newReimbursement,
		ipAddress,
		requestID,
	)

	return newReimbursement, nil
}
