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

// --- Test Suite Setup for OvertimeRepository ---

type OvertimeRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo OvertimeRepository
}

// SetupSuite runs before the tests in the suite are run.
func (s *OvertimeRepositorySuite) SetupSuite() {
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
	s.repo = NewOvertimeGormRepository(db)
}

// TearDownTest runs after each test in the suite.
func (s *OvertimeRepositorySuite) TearDownTest() {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

// TestOvertimeRepository runs the test suite.
func TestOvertimeRepository(t *testing.T) {
	suite.Run(t, new(OvertimeRepositorySuite))
}

// --- Test Cases ---

func (s *OvertimeRepositorySuite) TestCreateOvertime() {
	overtimeID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	testCases := []struct {
		name     string
		overtime *domain.Overtime
		mock     func()
		wantErr  bool
	}{
		{
			name: "Success",
			overtime: &domain.Overtime{
				BaseModel: domain.BaseModel{ID: overtimeID},
				UserID:    userID,
				Date:      now,
				Hours:     2.5,
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "overtimes" ("created_at","updated_at","deleted_at","created_by","updated_by","ip_address","user_id","date","hours","payroll_period_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), userID, now, 2.5, nil, overtimeID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(overtimeID))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			overtime: &domain.Overtime{
				BaseModel: domain.BaseModel{ID: overtimeID},
				UserID:    userID,
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "overtimes"`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			createdOvertime, err := s.repo.CreateOvertime(tc.overtime)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.overtime.ID, createdOvertime.ID)
			}
		})
	}
}

func (s *OvertimeRepositorySuite) TestGetOvertimeByID() {
	overtimeID := uuid.New()

	testCases := []struct {
		name    string
		id      uuid.UUID
		mock    func()
		wantErr bool
		wantNil bool
	}{
		{
			name: "Success",
			id:   overtimeID,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(overtimeID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "overtimes" WHERE "overtimes"."id" = $1 AND "overtimes"."deleted_at" IS NULL ORDER BY "overtimes"."id" LIMIT $2`)).
					WithArgs(overtimeID, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			id:   overtimeID,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "overtimes" WHERE "overtimes"."id" = $1 AND "overtimes"."deleted_at" IS NULL ORDER BY "overtimes"."id" LIMIT $2`)).
					WithArgs(overtimeID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
		{
			name: "DB Error",
			id:   overtimeID,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "overtimes" WHERE "overtimes"."id" = $1 AND "overtimes"."deleted_at" IS NULL ORDER BY "overtimes"."id" LIMIT $2`)).
					WithArgs(overtimeID, 1).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
			wantNil: false,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			overtime, err := s.repo.GetOvertimeByID(tc.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, overtime)
			} else if !tc.wantErr {
				assert.NotNil(t, overtime)
			}
		})
	}
}

func (s *OvertimeRepositorySuite) TestGetOvertimeByUserIDAndDate() {
	userID := uuid.New()
	date := time.Now()
	dateStr := date.Format("2006-01-02")

	testCases := []struct {
		name    string
		mock    func()
		wantLen int
		wantErr bool
	}{
		{
			name: "Success - Found Records",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"}).
					AddRow(uuid.New(), userID).
					AddRow(uuid.New(), userID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "overtimes" WHERE (user_id = $1 AND date = $2) AND "overtimes"."deleted_at" IS NULL`)).
					WithArgs(userID, dateStr).
					WillReturnRows(rows)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "overtimes" WHERE (user_id = $1 AND date = $2) AND "overtimes"."deleted_at" IS NULL`)).
					WithArgs(userID, dateStr).
					WillReturnError(errors.New("db error"))
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			overtimes, err := s.repo.GetOvertimeByUserIDAndDate(userID, date)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, overtimes, tc.wantLen)
			}
		})
	}
}

func (s *OvertimeRepositorySuite) TestGetOvertimesByUserIDAndPeriod() {
	userID := uuid.New()
	startDate := time.Now().Add(-5 * 24 * time.Hour)
	endDate := time.Now()
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	testCases := []struct {
		name    string
		mock    func()
		wantLen int
		wantErr bool
	}{
		{
			name: "Success - Found Records",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"}).
					AddRow(uuid.New(), userID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "overtimes" WHERE (user_id = $1 AND date >= $2 AND date <= $3) AND "overtimes"."deleted_at" IS NULL`)).
					WithArgs(userID, startDateStr, endDateStr).
					WillReturnRows(rows)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "overtimes" WHERE (user_id = $1 AND date >= $2 AND date <= $3) AND "overtimes"."deleted_at" IS NULL`)).
					WithArgs(userID, startDateStr, endDateStr).
					WillReturnError(errors.New("db error"))
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			overtimes, err := s.repo.GetOvertimesByUserIDAndPeriod(userID, startDate, endDate)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, overtimes, tc.wantLen)
			}
		})
	}
}

func (s *OvertimeRepositorySuite) TestGetOvertimesByUserIDAndPayrollPeriodID() {
	userID := uuid.New()
	payrollPeriodID := uuid.New()

	testCases := []struct {
		name    string
		mock    func()
		wantLen int
		wantErr bool
	}{
		{
			name: "Success - Found Records",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"}).
					AddRow(uuid.New(), userID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "overtimes" WHERE (user_id = $1 AND payroll_period_id = $2) AND "overtimes"."deleted_at" IS NULL`)).
					WithArgs(userID, payrollPeriodID).
					WillReturnRows(rows)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "overtimes" WHERE (user_id = $1 AND payroll_period_id = $2) AND "overtimes"."deleted_at" IS NULL`)).
					WithArgs(userID, payrollPeriodID).
					WillReturnError(errors.New("db error"))
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			overtimes, err := s.repo.GetOvertimesByUserIDAndPayrollPeriodID(userID, payrollPeriodID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, overtimes, tc.wantLen)
			}
		})
	}
}

func (s *OvertimeRepositorySuite) TestUpdateOvertime() {
	overtime := &domain.Overtime{
		BaseModel: domain.BaseModel{ID: uuid.New()},
	}

	testCases := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "Success",
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "overtimes" SET`)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "overtimes" SET`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.UpdateOvertime(overtime)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *OvertimeRepositorySuite) TestUpdateOvertimesTx() {
	overtimes := []domain.Overtime{
		{BaseModel: domain.BaseModel{ID: uuid.New()}},
	}

	testCases := []struct {
		name     string
		mock     func()
		wantErr  bool
		useNilTx bool
	}{
		{
			name: "Success",
			mock: func() {
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "overtimes" SET`)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "overtimes" SET`)).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:     "Nil Transaction",
			mock:     func() {},
			wantErr:  true,
			useNilTx: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.useNilTx {
				err := s.repo.UpdateOvertimesTx(nil, overtimes)
				assert.Error(t, err)
				return
			}

			s.mock.ExpectBegin()
			tc.mock()
			if tc.wantErr {
				s.mock.ExpectRollback()
			} else {
				s.mock.ExpectCommit()
			}

			err := s.db.Transaction(func(tx *gorm.DB) error {
				return s.repo.UpdateOvertimesTx(tx, overtimes)
			})

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
