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

// --- Test Suite Setup for ReimbursementRepository ---

type ReimbursementRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo ReimbursementRepository
}

// SetupSuite runs before the tests in the suite are run.
func (s *ReimbursementRepositorySuite) SetupSuite() {
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
	s.repo = NewReimbursementGormRepository(db)
}

// TearDownTest runs after each test in the suite.
func (s *ReimbursementRepositorySuite) TearDownTest() {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

// TestReimbursementRepository runs the test suite.
func TestReimbursementRepository(t *testing.T) {
	suite.Run(t, new(ReimbursementRepositorySuite))
}

// --- Test Cases ---

func (s *ReimbursementRepositorySuite) TestCreateReimbursement() {
	reimbursementID := uuid.New()
	userID := uuid.New()

	testCases := []struct {
		name          string
		reimbursement *domain.Reimbursement
		mock          func()
		wantErr       bool
	}{
		{
			name: "Success",
			reimbursement: &domain.Reimbursement{
				BaseModel: domain.BaseModel{ID: reimbursementID},
				UserID:    userID,
				Amount:    100.50,
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "reimbursements"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(reimbursementID))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			reimbursement: &domain.Reimbursement{
				BaseModel: domain.BaseModel{ID: reimbursementID},
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "reimbursements"`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.CreateReimbursement(tc.reimbursement)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ReimbursementRepositorySuite) TestGetReimbursementByID() {
	reimbursementID := uuid.New()

	testCases := []struct {
		name    string
		id      uuid.UUID
		mock    func()
		wantErr bool
		wantNil bool
	}{
		{
			name: "Success",
			id:   reimbursementID,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(reimbursementID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reimbursements" WHERE "reimbursements"."id" = $1 AND "reimbursements"."deleted_at" IS NULL ORDER BY "reimbursements"."id" LIMIT $2`)).
					WithArgs(reimbursementID, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			id:   reimbursementID,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reimbursements" WHERE "reimbursements"."id" = $1 AND "reimbursements"."deleted_at" IS NULL ORDER BY "reimbursements"."id" LIMIT $2`)).
					WithArgs(reimbursementID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			reimbursement, err := s.repo.GetReimbursementByID(tc.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, reimbursement)
			} else {
				assert.NotNil(t, reimbursement)
			}
		})
	}
}

func (s *ReimbursementRepositorySuite) TestGetReimbursementsByUserIDAndPeriod() {
	userID := uuid.New()
	startDate := time.Now().Add(-30 * 24 * time.Hour)
	endDate := time.Now()

	testCases := []struct {
		name    string
		mock    func()
		wantErr bool
		wantLen int
	}{
		{
			name: "Success",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"}).
					AddRow(uuid.New(), userID).
					AddRow(uuid.New(), userID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reimbursements" WHERE (user_id = $1 AND created_at >= $2 AND created_at <= $3) AND "reimbursements"."deleted_at" IS NULL`)).
					WithArgs(userID, startDate, endDate).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reimbursements" WHERE (user_id = $1 AND created_at >= $2 AND created_at <= $3) AND "reimbursements"."deleted_at" IS NULL`)).
					WithArgs(userID, startDate, endDate).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			reimbursements, err := s.repo.GetReimbursementsByUserIDAndPeriod(userID, startDate, endDate)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, reimbursements, tc.wantLen)
			}
		})
	}
}

func (s *ReimbursementRepositorySuite) TestUpdateReimbursement() {
	reimbursement := &domain.Reimbursement{
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
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "reimbursements" SET`)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "reimbursements" SET`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.UpdateReimbursement(reimbursement)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ReimbursementRepositorySuite) TestUpdateReimbursementsTx() {
	reimbursements := []domain.Reimbursement{
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
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "reimbursements" SET`)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "reimbursements" SET`)).
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
				err := s.repo.UpdateReimbursementsTx(nil, reimbursements)
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
				return s.repo.UpdateReimbursementsTx(tx, reimbursements)
			})

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
