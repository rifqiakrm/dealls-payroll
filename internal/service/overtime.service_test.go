package service_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
	mockRepo "payroll-system/tests/mocks/repository"
)

func TestOvertimeService_SubmitOvertime(t *testing.T) {
	userID := uuid.New()
	ip := "127.0.0.1"
	requestID := "req-123"
	date := time.Date(2025, 8, 18, 0, 0, 0, 0, time.UTC) // Monday

	tests := []struct {
		name             string
		hours            float64
		mockExisting     []domain.Overtime
		mockGetError     error
		mockCreateError  error
		expectedError    string
		expectCreateCall bool
	}{
		{
			name:             "successful submission with no existing overtime",
			hours:            2.0,
			mockExisting:     []domain.Overtime{},
			expectCreateCall: true,
		},
		{
			name:         "exceed daily max hours",
			hours:        2.5,
			mockExisting: []domain.Overtime{{Hours: 1.0}},
			expectedError: fmt.Sprintf(
				"total overtime hours for %s cannot exceed %.1f hours",
				date.Format("2006-01-02"), service.MaxOvertimeHoursPerDay),
		},
		{
			name:          "get overtime repo error",
			hours:         2.0,
			mockGetError:  errors.New("db error"),
			expectedError: "db error",
		},
		{
			name:             "create overtime repo error",
			hours:            1.0,
			mockExisting:     []domain.Overtime{},
			mockCreateError:  errors.New("create failed"),
			expectedError:    "create failed",
			expectCreateCall: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOvertimeRepo := mockRepo.NewMockOvertimeRepository(ctrl)
			mockAuditRepo := mockRepo.NewMockAuditLogRepository(ctrl)
			svc := service.NewOvertimeService(mockOvertimeRepo, mockAuditRepo)

			// Mock GetOvertimeByUserIDAndDate
			mockOvertimeRepo.
				EXPECT().
				GetOvertimeByUserIDAndDate(userID, date).
				Return(tt.mockExisting, tt.mockGetError).
				AnyTimes()

			if tt.expectCreateCall {
				mockOvertimeRepo.
					EXPECT().
					CreateOvertime(gomock.Any()).
					Return(&domain.Overtime{}, tt.mockCreateError).
					Times(1)
			}

			// Audit log can always be called
			mockAuditRepo.
				EXPECT().
				Create(gomock.Any()).
				Return(nil).
				AnyTimes()

			ot, err := svc.SubmitOvertime(userID, date, tt.hours, ip, requestID)

			if tt.expectedError != "" {
				assert.Nil(t, ot)
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NotNil(t, ot)
				assert.NoError(t, err)
				assert.Equal(t, userID, ot.UserID)
				assert.Equal(t, tt.hours, ot.Hours)
			}
		})
	}
}
