package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
	mockRepo "payroll-system/tests/mocks/repository"
)

func TestSubmitAttendance(t *testing.T) {
	userID := uuid.New()
	ip := "127.0.0.1"
	requestID := "req-123"
	now := time.Date(2025, 8, 18, 9, 0, 0, 0, time.UTC) // Monday

	tests := []struct {
		name            string
		checkIn         time.Time
		checkOut        time.Time
		mockExisting    *domain.Attendance
		mockGetError    error
		mockCreateError error
		mockUpdateError error
		expectedError   string
		expectCreate    bool
		expectUpdate    bool
	}{
		{
			name:          "weekend submission error",
			checkIn:       time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC), // Saturday
			checkOut:      time.Date(2025, 8, 16, 17, 0, 0, 0, time.UTC),
			expectedError: "attendance cannot be submitted on weekends",
		},
		{
			name:         "new attendance success",
			checkIn:      now,
			checkOut:     now.Add(8 * time.Hour),
			mockExisting: nil,
			expectCreate: true,
		},
		{
			name:     "update existing attendance success",
			checkIn:  now,
			checkOut: now.Add(8 * time.Hour),
			mockExisting: &domain.Attendance{
				BaseModel:    domain.BaseModel{ID: uuid.New()},
				UserID:       userID,
				Date:         now, // must match checkIn to trigger update
				CheckInTime:  now.Add(-1 * time.Hour),
				CheckOutTime: now.Add(7 * time.Hour),
			},
			expectUpdate: true,
		},
		{
			name:          "get attendance error",
			checkIn:       now,
			checkOut:      now.Add(8 * time.Hour),
			mockGetError:  errors.New("db error"),
			expectedError: "db error",
		},
		{
			name:            "create attendance error",
			checkIn:         now,
			checkOut:        now.Add(8 * time.Hour),
			mockExisting:    nil,
			mockCreateError: errors.New("create failed"),
			expectedError:   "create failed",
			expectCreate:    true,
		},
		{
			name:     "update attendance error",
			checkIn:  now,
			checkOut: now.Add(8 * time.Hour),
			mockExisting: &domain.Attendance{
				BaseModel: domain.BaseModel{ID: uuid.New()},
				UserID:    userID,
				Date:      now,
			},
			mockUpdateError: errors.New("update failed"),
			expectedError:   "update failed",
			expectUpdate:    true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAttendanceRepo := mockRepo.NewMockAttendanceRepository(ctrl)
			mockAuditRepo := mockRepo.NewMockAuditLogRepository(ctrl)
			svc := service.NewAttendanceService(mockAttendanceRepo, mockAuditRepo)

			// Mock GetAttendanceByUserIDAndDate
			mockAttendanceRepo.
				EXPECT().
				GetAttendanceByUserIDAndDate(userID, tt.checkIn).
				Return(tt.mockExisting, tt.mockGetError).
				AnyTimes()

			if tt.expectCreate {
				mockAttendanceRepo.
					EXPECT().
					CreateAttendance(gomock.Any()).
					Return(tt.mockCreateError).
					Times(1)
			}

			if tt.expectUpdate {
				mockAttendanceRepo.
					EXPECT().
					UpdateAttendance(gomock.Any()).
					Return(tt.mockUpdateError).
					Times(1)
			}

			// Audit log can always be called
			mockAuditRepo.
				EXPECT().
				Create(gomock.Any()).
				Return(nil).
				AnyTimes()

			att, err := svc.SubmitAttendance(userID, tt.checkIn, tt.checkOut, ip, requestID)

			if tt.expectedError != "" {
				assert.Nil(t, att)
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NotNil(t, att)
				assert.NoError(t, err)
				assert.Equal(t, userID, att.UserID)
			}
		})
	}
}
