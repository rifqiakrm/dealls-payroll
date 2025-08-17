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

func TestPayrollPeriodService_CreatePayrollPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPayrollRepo := mockRepo.NewMockPayrollPeriodRepository(ctrl)
	mockAuditRepo := mockRepo.NewMockAuditLogRepository(ctrl)
	svc := service.NewPayrollPeriodService(mockPayrollRepo, mockAuditRepo)

	createdBy := uuid.New()
	ip := "127.0.0.1"
	requestID := "req-123"
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 7)

	tests := []struct {
		name               string
		startDate          time.Time
		endDate            time.Time
		setupMocks         func()
		expectedErr        string
		expectPeriodNotNil bool
	}{
		{
			name:      "success",
			startDate: startDate,
			endDate:   endDate,
			setupMocks: func() {
				mockPayrollRepo.EXPECT().
					GetOverlappingPayrollPeriods(startDate, endDate).
					Return([]domain.PayrollPeriod{}, nil).
					Times(1)
				mockPayrollRepo.EXPECT().
					CreatePayrollPeriod(gomock.Any()).
					Return(nil).
					Times(1)
				mockAuditRepo.EXPECT().
					Create(gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedErr:        "",
			expectPeriodNotNil: true,
		},
		{
			name:      "invalid dates",
			startDate: endDate,
			endDate:   startDate,
			setupMocks: func() {
				// no repo calls expected
			},
			expectedErr:        "end date must be after start date",
			expectPeriodNotNil: false,
		},
		{
			name:      "overlapping period",
			startDate: startDate,
			endDate:   endDate,
			setupMocks: func() {
				mockPayrollRepo.EXPECT().
					GetOverlappingPayrollPeriods(startDate, endDate).
					Return([]domain.PayrollPeriod{{}}, nil).
					Times(1)
			},
			expectedErr:        "a payroll period overlapping with these dates already exists",
			expectPeriodNotNil: false,
		},
		{
			name:      "repository error",
			startDate: startDate,
			endDate:   endDate,
			setupMocks: func() {
				mockPayrollRepo.EXPECT().
					GetOverlappingPayrollPeriods(startDate, endDate).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			expectedErr:        "db error",
			expectPeriodNotNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			period, err := svc.CreatePayrollPeriod(tt.startDate, tt.endDate, createdBy, ip, requestID)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				assert.Nil(t, period)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, period)
			}
		})
	}
}

func TestPayrollPeriodService_MarkPayrollPeriodAsProcessed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPayrollRepo := mockRepo.NewMockPayrollPeriodRepository(ctrl)
	mockAuditRepo := mockRepo.NewMockAuditLogRepository(ctrl)
	svc := service.NewPayrollPeriodService(mockPayrollRepo, mockAuditRepo)

	periodID := uuid.New()
	updatedBy := uuid.New()
	ip := "127.0.0.1"

	tests := []struct {
		name        string
		setupMocks  func()
		expectedErr string
	}{
		{
			name: "success",
			setupMocks: func() {
				mockPayrollRepo.EXPECT().
					GetPayrollPeriodByID(periodID).
					Return(&domain.PayrollPeriod{BaseModel: domain.BaseModel{ID: periodID}}, nil).
					Times(1)
				mockPayrollRepo.EXPECT().
					MarkPayrollPeriodAsProcessed(periodID).
					Return(nil).
					Times(1)
				mockAuditRepo.EXPECT().
					Create(gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedErr: "",
		},
		{
			name: "period not found",
			setupMocks: func() {
				mockPayrollRepo.EXPECT().
					GetPayrollPeriodByID(periodID).
					Return(nil, nil).
					Times(1)
			},
			expectedErr: "payroll period not found",
		},
		{
			name: "repository error",
			setupMocks: func() {
				mockPayrollRepo.EXPECT().
					GetPayrollPeriodByID(periodID).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			expectedErr: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := svc.MarkPayrollPeriodAsProcessed(periodID, updatedBy, ip)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
