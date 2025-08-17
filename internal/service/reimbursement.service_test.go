package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"payroll-system/internal/service"
	mockRepo "payroll-system/tests/mocks/repository"
)

func TestReimbursementService_SubmitReimbursement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReimbursementRepo := mockRepo.NewMockReimbursementRepository(ctrl)
	mockAuditRepo := mockRepo.NewMockAuditLogRepository(ctrl)

	svc := service.NewReimbursementService(mockReimbursementRepo, mockAuditRepo)

	userID := uuid.New()
	ipAddress := "127.0.0.1"
	requestID := uuid.New().String()
	description := "Travel expense"
	amount := 100.0

	tests := []struct {
		name       string
		setupMocks func()
		expectErr  string
	}{
		{
			name: "success",
			setupMocks: func() {
				mockReimbursementRepo.EXPECT().CreateReimbursement(gomock.Any()).Return(nil)
				mockAuditRepo.EXPECT().
					Create(gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectErr: "",
		},
		{
			name: "reimbursement repo error",
			setupMocks: func() {
				mockReimbursementRepo.EXPECT().CreateReimbursement(gomock.Any()).Return(errors.New("db error"))
			},
			expectErr: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			reimbursement, err := svc.SubmitReimbursement(userID, amount, description, ipAddress, requestID)
			if tt.expectErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectErr, err.Error())
				assert.Nil(t, reimbursement)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reimbursement)
				assert.Equal(t, userID, reimbursement.UserID)
				assert.Equal(t, amount, reimbursement.Amount)
				assert.Equal(t, description, reimbursement.Description)
				// Approximate check for timestamps
				assert.WithinDuration(t, time.Now(), reimbursement.CreatedAt, 2*time.Second)
			}
		})
	}
}
