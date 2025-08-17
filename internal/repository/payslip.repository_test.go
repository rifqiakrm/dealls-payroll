package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// --- Test Suite Setup for PayslipRepository ---

type PayslipRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo PayslipRepository
}

// SetupSuite runs before the tests in the suite are run.
func (s *PayslipRepositorySuite) SetupSuite() {
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
	s.repo = NewPayslipGormRepository(db)
}

// TearDownTest runs after each test in the suite.
func (s *PayslipRepositorySuite) TearDownTest() {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

// TestPayslipRepository runs the test suite.
func TestPayslipRepository(t *testing.T) {
	suite.Run(t, new(PayslipRepositorySuite))
}

// --- Test Cases ---

func (s *PayslipRepositorySuite) TestCreatePayslip() {
	payslipID := uuid.New()
	userID := uuid.New()
	periodID := uuid.New()

	testCases := []struct {
		name    string
		payslip *domain.Payslip
		mock    func()
		wantErr bool
	}{
		{
			name: "Success",
			payslip: &domain.Payslip{
				BaseModel:       domain.BaseModel{ID: payslipID},
				UserID:          userID,
				PayrollPeriodID: periodID,
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payslips"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(payslipID))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			payslip: &domain.Payslip{
				BaseModel: domain.BaseModel{ID: payslipID},
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payslips"`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.CreatePayslip(tc.payslip)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *PayslipRepositorySuite) TestGetPayslipByID() {
	payslipID := uuid.New()

	testCases := []struct {
		name    string
		id      uuid.UUID
		mock    func()
		wantErr bool
		wantNil bool
	}{
		{
			name: "Success",
			id:   payslipID,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(payslipID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payslips" WHERE "payslips"."id" = $1 AND "payslips"."deleted_at" IS NULL ORDER BY "payslips"."id" LIMIT $2`)).
					WithArgs(payslipID, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			id:   payslipID,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payslips" WHERE "payslips"."id" = $1 AND "payslips"."deleted_at" IS NULL ORDER BY "payslips"."id" LIMIT $2`)).
					WithArgs(payslipID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			payslip, err := s.repo.GetPayslipByID(tc.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, payslip)
			} else {
				assert.NotNil(t, payslip)
			}
		})
	}
}

func (s *PayslipRepositorySuite) TestGetPayslipByUserIDAndPeriodID() {
	userID := uuid.New()
	periodID := uuid.New()

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
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payslips" WHERE (user_id = $1 AND payroll_period_id = $2) AND "payslips"."deleted_at" IS NULL ORDER BY "payslips"."id" LIMIT $3`)).
					WithArgs(userID, periodID, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payslips" WHERE (user_id = $1 AND payroll_period_id = $2) AND "payslips"."deleted_at" IS NULL ORDER BY "payslips"."id" LIMIT $3`)).
					WithArgs(userID, periodID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			payslip, err := s.repo.GetPayslipByUserIDAndPeriodID(userID, periodID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, payslip)
			} else {
				assert.NotNil(t, payslip)
			}
		})
	}
}

func (s *PayslipRepositorySuite) TestGetAllPayslipsByPeriodID() {
	periodID := uuid.New()

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
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payslips" WHERE payroll_period_id = $1 AND "payslips"."deleted_at" IS NULL`)).
					WithArgs(periodID).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payslips" WHERE payroll_period_id = $1 AND "payslips"."deleted_at" IS NULL`)).
					WithArgs(periodID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			payslips, err := s.repo.GetAllPayslipsByPeriodID(periodID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, payslips, tc.wantLen)
			}
		})
	}
}

func (s *PayslipRepositorySuite) TestCreatePayslipTx() {
	payslip := &domain.Payslip{BaseModel: domain.BaseModel{ID: uuid.New()}}

	testCases := []struct {
		name     string
		mock     func()
		wantErr  bool
		useNilTx bool
	}{
		{
			name: "Success",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payslips"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(payslip.ID))
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payslips"`)).
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
				err := s.repo.CreatePayslipTx(nil, payslip)
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
				return s.repo.CreatePayslipTx(tx, payslip)
			})

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
