package service_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
	mockrepo "payroll-system/tests/mocks/repository"

	"go.uber.org/mock/gomock"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	cleanup := func() { db.Close() }
	return gormDB, mock, cleanup
}

func TestRunPayroll(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(
			t *testing.T,
			payslipRepo *mockrepo.MockPayslipRepository,
			payrollPeriodRepo *mockrepo.MockPayrollPeriodRepository,
			employeeProfileRepo *mockrepo.MockEmployeeProfileRepository,
			attendanceRepo *mockrepo.MockAttendanceRepository,
			overtimeRepo *mockrepo.MockOvertimeRepository,
			reimbursementRepo *mockrepo.MockReimbursementRepository,
			auditRepo *mockrepo.MockAuditLogRepository,
		)
		expectError bool
	}{
		{
			name: "success run payroll",
			mockSetup: func(t *testing.T, payslipRepo *mockrepo.MockPayslipRepository, payrollPeriodRepo *mockrepo.MockPayrollPeriodRepository,
				employeeProfileRepo *mockrepo.MockEmployeeProfileRepository, attendanceRepo *mockrepo.MockAttendanceRepository,
				overtimeRepo *mockrepo.MockOvertimeRepository, reimbursementRepo *mockrepo.MockReimbursementRepository, auditRepo *mockrepo.MockAuditLogRepository) {

				now := time.Now()
				userID := uuid.New()

				// Payroll period exists and not processed
				payrollPeriodRepo.EXPECT().
					GetPayrollPeriodByID(gomock.Any()).
					Return(&domain.PayrollPeriod{
						BaseModel:   domain.BaseModel{ID: uuid.New()},
						StartDate:   now.Add(-10 * 24 * time.Hour),
						EndDate:     now,
						IsProcessed: false,
					}, nil)

				// Employees
				employeeProfileRepo.EXPECT().
					GetAllEmployeeProfiles().
					Return([]domain.EmployeeProfile{
						{
							UserID: userID,
							Salary: 1000,
						},
					}, nil)

				// CalculatePayslip repo calls
				employeeProfileRepo.EXPECT().
					GetEmployeeProfileByUserID(gomock.Any()).
					Return(&domain.EmployeeProfile{
						UserID: userID,
						Salary: 1000,
					}, nil)

				attendanceRepo.EXPECT().
					GetAttendancesByUserIDAndPeriod(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]domain.Attendance{
						{
							CheckInTime:  now.Add(-8 * time.Hour),
							CheckOutTime: now,
							Date:         now,
						},
					}, nil)

				overtimeRepo.EXPECT().
					GetOvertimesByUserIDAndPeriod(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]domain.Overtime{}, nil)

				reimbursementRepo.EXPECT().
					GetReimbursementsByUserIDAndPeriod(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]domain.Reimbursement{}, nil)

				// Save payslip and related records
				payslipRepo.EXPECT().
					CreatePayslipTx(gomock.Any(), gomock.Any()).
					Return(nil)
				attendanceRepo.EXPECT().UpdateAttendancesTx(gomock.Any(), gomock.Any()).Return(nil)
				overtimeRepo.EXPECT().UpdateOvertimesTx(gomock.Any(), gomock.Any()).Return(nil)
				reimbursementRepo.EXPECT().UpdateReimbursementsTx(gomock.Any(), gomock.Any()).Return(nil)

				// Mark payroll as processed
				payrollPeriodRepo.EXPECT().
					MarkPayrollPeriodAsProcessedTx(gomock.Any(), gomock.Any()).
					Return(nil)

				// Audit logs
				auditRepo.EXPECT().
					Create(gomock.Any()).
					Return(nil).
					AnyTimes()
			},
			expectError: false,
		},
		{
			name: "payroll period not found",
			mockSetup: func(t *testing.T, payslipRepo *mockrepo.MockPayslipRepository, payrollPeriodRepo *mockrepo.MockPayrollPeriodRepository,
				employeeProfileRepo *mockrepo.MockEmployeeProfileRepository, attendanceRepo *mockrepo.MockAttendanceRepository,
				overtimeRepo *mockrepo.MockOvertimeRepository, reimbursementRepo *mockrepo.MockReimbursementRepository, auditRepo *mockrepo.MockAuditLogRepository) {

				payrollPeriodRepo.EXPECT().
					GetPayrollPeriodByID(gomock.Any()).
					Return(nil, nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Mocks
			payslipRepo := mockrepo.NewMockPayslipRepository(ctrl)
			payrollPeriodRepo := mockrepo.NewMockPayrollPeriodRepository(ctrl)
			employeeProfileRepo := mockrepo.NewMockEmployeeProfileRepository(ctrl)
			attendanceRepo := mockrepo.NewMockAttendanceRepository(ctrl)
			overtimeRepo := mockrepo.NewMockOvertimeRepository(ctrl)
			reimbursementRepo := mockrepo.NewMockReimbursementRepository(ctrl)
			auditRepo := mockrepo.NewMockAuditLogRepository(ctrl)

			// Setup DB
			db, sqlmock, cleanup := setupTestDB(t)
			defer cleanup()

			// Begin transaction for all cases
			sqlmock.ExpectBegin()
			if !tt.expectError {
				sqlmock.ExpectCommit()
			} else {
				sqlmock.ExpectRollback()
			}

			// Setup mocks
			if tt.mockSetup != nil {
				tt.mockSetup(t, payslipRepo, payrollPeriodRepo, employeeProfileRepo, attendanceRepo, overtimeRepo, reimbursementRepo, auditRepo)
			}

			svc := service.NewPayrollService(payslipRepo, payrollPeriodRepo, employeeProfileRepo, attendanceRepo, overtimeRepo, reimbursementRepo, auditRepo, db)

			err := svc.RunPayroll(uuid.New(), uuid.New(), "127.0.0.1", "req-123")
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Ensure all sqlmock expectations are met
			require.NoError(t, sqlmock.ExpectationsWereMet())
		})
	}
}
