package service_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
	mockRepo "payroll-system/tests/mocks/repository"
)

func TestPayslipService_GetEmployeePayslip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPayslipRepo := mockRepo.NewMockPayslipRepository(ctrl)
	mockPeriodRepo := mockRepo.NewMockPayrollPeriodRepository(ctrl)
	mockAttendanceRepo := mockRepo.NewMockAttendanceRepository(ctrl)
	mockOvertimeRepo := mockRepo.NewMockOvertimeRepository(ctrl)

	svc := service.NewPayslipService(mockPayslipRepo, mockPeriodRepo, mockAttendanceRepo, mockOvertimeRepo)

	userID := uuid.New()
	periodID := uuid.New()

	tests := []struct {
		name       string
		setupMocks func()
		expectErr  string
	}{
		{
			name: "success",
			setupMocks: func() {
				period := &domain.PayrollPeriod{BaseModel: domain.BaseModel{ID: periodID}, IsProcessed: true}
				payslip := &domain.Payslip{UserID: userID, PayrollPeriodID: periodID, TotalTakeHomePay: 1000}
				mockPeriodRepo.EXPECT().GetPayrollPeriodByID(periodID).Return(period, nil)
				mockPayslipRepo.EXPECT().GetPayslipByUserIDAndPeriodID(userID, periodID).Return(payslip, nil)
				mockAttendanceRepo.EXPECT().GetAttendancesByUserIDAndPayrollPeriodID(userID, periodID).Return(nil, nil)
				mockOvertimeRepo.EXPECT().GetOvertimesByUserIDAndPayrollPeriodID(userID, periodID).Return(nil, nil)
			},
			expectErr: "",
		},
		{
			name: "payroll period not found",
			setupMocks: func() {
				mockPeriodRepo.EXPECT().GetPayrollPeriodByID(periodID).Return(nil, nil)
			},
			expectErr: "payroll period not found",
		},
		{
			name: "payroll not processed",
			setupMocks: func() {
				mockPeriodRepo.EXPECT().GetPayrollPeriodByID(periodID).Return(&domain.PayrollPeriod{IsProcessed: false}, nil)
			},
			expectErr: "payslip can only be generated for processed payroll periods",
		},
		{
			name: "payslip not found",
			setupMocks: func() {
				mockPeriodRepo.EXPECT().GetPayrollPeriodByID(periodID).Return(&domain.PayrollPeriod{IsProcessed: true}, nil)
				mockPayslipRepo.EXPECT().GetPayslipByUserIDAndPeriodID(userID, periodID).Return(nil, nil)
			},
			expectErr: "payslip not found for this user and period",
		},
		{
			name: "repo error",
			setupMocks: func() {
				mockPeriodRepo.EXPECT().GetPayrollPeriodByID(periodID).Return(nil, errors.New("db error"))
			},
			expectErr: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			payslip, err := svc.GetEmployeePayslip(userID, periodID)
			if tt.expectErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectErr, err.Error())
				assert.Nil(t, payslip)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payslip)
			}
		})
	}
}

func TestPayslipService_GetPayslipSummaryForPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPayslipRepo := mockRepo.NewMockPayslipRepository(ctrl)
	mockPeriodRepo := mockRepo.NewMockPayrollPeriodRepository(ctrl)
	mockAttendanceRepo := mockRepo.NewMockAttendanceRepository(ctrl)
	mockOvertimeRepo := mockRepo.NewMockOvertimeRepository(ctrl)

	svc := service.NewPayslipService(mockPayslipRepo, mockPeriodRepo, mockAttendanceRepo, mockOvertimeRepo)

	periodID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		setupMocks func()
		expectErr  string
	}{
		{
			name: "success",
			setupMocks: func() {
				period := &domain.PayrollPeriod{BaseModel: domain.BaseModel{ID: periodID}, IsProcessed: true}
				payslips := []domain.Payslip{
					{UserID: userID, PayrollPeriodID: periodID, TotalTakeHomePay: 1000},
				}
				mockPeriodRepo.EXPECT().GetPayrollPeriodByID(periodID).Return(period, nil)
				mockPayslipRepo.EXPECT().GetAllPayslipsByPeriodID(periodID).Return(payslips, nil)
				mockAttendanceRepo.EXPECT().GetAttendancesByUserIDAndPayrollPeriodID(userID, periodID).Return(nil, nil)
				mockOvertimeRepo.EXPECT().GetOvertimesByUserIDAndPayrollPeriodID(userID, periodID).Return(nil, nil)
			},
			expectErr: "",
		},
		{
			name: "period not found",
			setupMocks: func() {
				mockPeriodRepo.EXPECT().GetPayrollPeriodByID(periodID).Return(nil, nil)
			},
			expectErr: "payroll period not found",
		},
		{
			name: "period not processed",
			setupMocks: func() {
				mockPeriodRepo.EXPECT().GetPayrollPeriodByID(periodID).Return(&domain.PayrollPeriod{IsProcessed: false}, nil)
			},
			expectErr: "payslip summary can only be generated for processed payroll periods",
		},
		{
			name: "repo error",
			setupMocks: func() {
				mockPeriodRepo.EXPECT().GetPayrollPeriodByID(periodID).Return(nil, errors.New("db error"))
			},
			expectErr: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			payslips, total, err := svc.GetPayslipSummaryForPeriod(periodID)
			if tt.expectErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectErr, err.Error())
				assert.Nil(t, payslips)
				assert.Equal(t, float64(0), total)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payslips)
				assert.Equal(t, 1000.0, total)
			}
		})
	}
}
