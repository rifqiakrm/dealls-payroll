package repository

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// --- Test Suite Setup for PayrollPeriodRepository ---

type PayrollPeriodRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo PayrollPeriodRepository
}

// SetupSuite runs before the tests in the suite are run.
func (s *PayrollPeriodRepositorySuite) SetupSuite() {
	sqlDB, mock, err := sqlmock.New()
	s.Require().NoError(err)

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	s.Require().NoError(err)

	s.db = db
	s.mock = mock
	s.repo = NewPayrollPeriodGormRepository(db)
}

// TearDownTest runs after each test in the suite.
func (s *PayrollPeriodRepositorySuite) TearDownTest() {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

// TestPayrollPeriodRepository runs the test suite.
func TestPayrollPeriodRepository(t *testing.T) {
	suite.Run(t, new(PayrollPeriodRepositorySuite))
}

// --- Test Cases ---

func (s *PayrollPeriodRepositorySuite) TestCreatePayrollPeriod() {
	periodID := uuid.New()
	startDate := time.Now()
	endDate := startDate.Add(14 * 24 * time.Hour)

	testCases := []struct {
		name    string
		period  *domain.PayrollPeriod
		mock    func()
		wantErr bool
	}{
		{
			name: "Success",
			period: &domain.PayrollPeriod{
				BaseModel: domain.BaseModel{ID: periodID},
				StartDate: startDate,
				EndDate:   endDate,
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payroll_periods" ("created_at","updated_at","deleted_at","created_by","updated_by","ip_address","start_date","end_date","is_processed","processed_at","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), startDate, endDate, false, nil, periodID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(periodID))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			period: &domain.PayrollPeriod{
				BaseModel: domain.BaseModel{ID: periodID},
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payroll_periods"`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.CreatePayrollPeriod(tc.period)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *PayrollPeriodRepositorySuite) TestGetPayrollPeriodByID() {
	periodID := uuid.New()

	testCases := []struct {
		name    string
		id      uuid.UUID
		mock    func()
		wantErr bool
		wantNil bool
	}{
		{
			name: "Success",
			id:   periodID,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(periodID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE "payroll_periods"."id" = $1 AND "payroll_periods"."deleted_at" IS NULL ORDER BY "payroll_periods"."id" LIMIT $2`)).
					WithArgs(periodID, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			id:   periodID,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE "payroll_periods"."id" = $1 AND "payroll_periods"."deleted_at" IS NULL ORDER BY "payroll_periods"."id" LIMIT $2`)).
					WithArgs(periodID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			period, err := s.repo.GetPayrollPeriodByID(tc.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, period)
			} else {
				assert.NotNil(t, period)
			}
		})
	}
}

func (s *PayrollPeriodRepositorySuite) TestGetActivePayrollPeriod() {
	testCases := []struct {
		name    string
		mock    func()
		wantErr bool
		wantNil bool
	}{
		{
			name: "Success",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(uuid.New())
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE is_processed = $1 AND "payroll_periods"."deleted_at" IS NULL ORDER BY start_date ASC,"payroll_periods"."id" LIMIT $2`)).
					WithArgs(false, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE is_processed = $1 AND "payroll_periods"."deleted_at" IS NULL ORDER BY start_date ASC,"payroll_periods"."id" LIMIT $2`)).
					WithArgs(false, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			period, err := s.repo.GetActivePayrollPeriod()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, period)
			} else {
				assert.NotNil(t, period)
			}
		})
	}
}

func (s *PayrollPeriodRepositorySuite) TestMarkPayrollPeriodAsProcessed() {
	periodID := uuid.New()

	testCases := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "Success",
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payroll_periods" SET "is_processed"=$1,"processed_at"=$2,"updated_at"=$3 WHERE id = $4 AND "payroll_periods"."deleted_at" IS NULL`)).
					WithArgs(true, sqlmock.AnyArg(), sqlmock.AnyArg(), periodID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payroll_periods" SET`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.MarkPayrollPeriodAsProcessed(periodID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *PayrollPeriodRepositorySuite) TestGetAllPayrollPeriods() {
	testCases := []struct {
		name    string
		mock    func()
		wantErr bool
		wantLen int
	}{
		{
			name: "Success",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()).AddRow(uuid.New())
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE "payroll_periods"."deleted_at" IS NULL ORDER BY start_date DESC`)).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE "payroll_periods"."deleted_at" IS NULL ORDER BY start_date DESC`)).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			periods, err := s.repo.GetAllPayrollPeriods()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, periods, tc.wantLen)
			}
		})
	}
}

func (s *PayrollPeriodRepositorySuite) TestGetPayrollPeriodByDates() {
	startDate := time.Now()
	endDate := startDate.Add(14 * 24 * time.Hour)

	testCases := []struct {
		name    string
		mock    func()
		wantErr bool
		wantNil bool
	}{
		{
			name: "Success",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(uuid.New())
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE (start_date = $1 AND end_date = $2) AND "payroll_periods"."deleted_at" IS NULL ORDER BY "payroll_periods"."id" LIMIT $3`)).
					WithArgs(startDate, endDate, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE (start_date = $1 AND end_date = $2) AND "payroll_periods"."deleted_at" IS NULL ORDER BY "payroll_periods"."id" LIMIT $3`)).
					WithArgs(startDate, endDate, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			period, err := s.repo.GetPayrollPeriodByDates(startDate, endDate)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, period)
			} else {
				assert.NotNil(t, period)
			}
		})
	}
}

func (s *PayrollPeriodRepositorySuite) TestMarkPayrollPeriodAsProcessedTx() {
	periodID := uuid.New()

	testCases := []struct {
		name    string
		mock    func()
		wantErr bool
		errMsg  string
	}{
		{
			name: "Success",
			mock: func() {
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payroll_periods" SET "is_processed"=$1,"processed_at"=$2,"updated_at"=$3 WHERE (id = $4 AND is_processed = $5) AND "payroll_periods"."deleted_at" IS NULL`)).
					WithArgs(true, sqlmock.AnyArg(), sqlmock.AnyArg(), periodID, false).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payroll_periods" SET "is_processed"=$1,"processed_at"=$2,"updated_at"=$3 WHERE (id = $4 AND is_processed = $5) AND "payroll_periods"."deleted_at" IS NULL`)).
					WithArgs(true, sqlmock.AnyArg(), sqlmock.AnyArg(), periodID, false).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
			errMsg:  "failed to mark payroll period as processed",
		},
		{
			name: "No Rows Affected",
			mock: func() {
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payroll_periods" SET "is_processed"=$1,"processed_at"=$2,"updated_at"=$3 WHERE (id = $4 AND is_processed = $5) AND "payroll_periods"."deleted_at" IS NULL`)).
					WithArgs(true, sqlmock.AnyArg(), sqlmock.AnyArg(), periodID, false).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errMsg:  "no payroll period updated",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.mock.ExpectBegin()
			tc.mock()
			if tc.wantErr {
				s.mock.ExpectRollback()
			} else {
				s.mock.ExpectCommit()
			}

			err := s.db.Transaction(func(tx *gorm.DB) error {
				return s.repo.MarkPayrollPeriodAsProcessedTx(tx, periodID)
			})

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *PayrollPeriodRepositorySuite) TestGetOverlappingPayrollPeriods() {
	startDate := time.Now()
	endDate := startDate.Add(14 * 24 * time.Hour)

	testCases := []struct {
		name    string
		mock    func()
		wantErr bool
		wantLen int
	}{
		{
			name: "Success",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(uuid.New())
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE (start_date <= $1 AND end_date >= $2) AND "payroll_periods"."deleted_at" IS NULL`)).
					WithArgs(endDate, startDate).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payroll_periods" WHERE (start_date <= $1 AND end_date >= $2) AND "payroll_periods"."deleted_at" IS NULL`)).
					WithArgs(endDate, startDate).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			periods, err := s.repo.GetOverlappingPayrollPeriods(startDate, endDate)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, periods, tc.wantLen)
			}
		})
	}
}
