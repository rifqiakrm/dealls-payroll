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

// --- Test Suite Setup for EmployeeProfileRepository ---

type EmployeeProfileRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo EmployeeProfileRepository
}

// SetupSuite runs before the tests in the suite are run.
func (s *EmployeeProfileRepositorySuite) SetupSuite() {
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
	s.repo = NewEmployeeProfileGormRepository(db)
}

// TearDownTest runs after each test in the suite.
func (s *EmployeeProfileRepositorySuite) TearDownTest() {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

// TestEmployeeProfileRepository runs the test suite.
func TestEmployeeProfileRepository(t *testing.T) {
	suite.Run(t, new(EmployeeProfileRepositorySuite))
}

// --- Test Cases ---

func (s *EmployeeProfileRepositorySuite) TestCreateEmployeeProfile() {
	profileID := uuid.New()
	userID := uuid.New()

	testCases := []struct {
		name    string
		profile *domain.EmployeeProfile
		mock    func()
		wantErr bool
	}{
		{
			name: "Success",
			profile: &domain.EmployeeProfile{
				BaseModel: domain.BaseModel{ID: profileID},
				UserID:    userID,
				Salary:    60000,
			},
			mock: func() {
				s.mock.ExpectBegin()
				// Corrected the SQL query and argument type for salary to float64.
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "employee_profiles" ("created_at","updated_at","deleted_at","created_by","updated_by","ip_address","user_id","salary","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), userID, float64(60000), profileID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(profileID))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			profile: &domain.EmployeeProfile{
				BaseModel: domain.BaseModel{ID: profileID},
				UserID:    userID,
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "employee_profiles"`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.CreateEmployeeProfile(tc.profile)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *EmployeeProfileRepositorySuite) TestGetEmployeeProfileByUserID() {
	userID := uuid.New()

	testCases := []struct {
		name    string
		userID  uuid.UUID
		mock    func()
		wantErr bool
		wantNil bool
	}{
		{
			name:   "Success",
			userID: userID,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"}).
					AddRow(uuid.New(), userID)
				// Corrected the query to handle LIMIT as a parameter.
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employee_profiles" WHERE user_id = $1 AND "employee_profiles"."deleted_at" IS NULL ORDER BY "employee_profiles"."id" LIMIT $2`)).
					WithArgs(userID, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:   "Not Found",
			userID: userID,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employee_profiles" WHERE user_id = $1 AND "employee_profiles"."deleted_at" IS NULL ORDER BY "employee_profiles"."id" LIMIT $2`)).
					WithArgs(userID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
		{
			name:   "DB Error",
			userID: userID,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employee_profiles" WHERE user_id = $1 AND "employee_profiles"."deleted_at" IS NULL ORDER BY "employee_profiles"."id" LIMIT $2`)).
					WithArgs(userID, 1).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
			wantNil: false,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			profile, err := s.repo.GetEmployeeProfileByUserID(tc.userID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.wantNil {
				assert.Nil(t, profile)
			} else if !tc.wantErr {
				assert.NotNil(t, profile)
				assert.Equal(t, tc.userID, profile.UserID)
			}
		})
	}
}

func (s *EmployeeProfileRepositorySuite) TestGetAllEmployeeProfiles() {
	testCases := []struct {
		name    string
		mock    func()
		wantLen int
		wantErr bool
	}{
		{
			name: "Success with multiple profiles",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"}).
					AddRow(uuid.New(), uuid.New()).
					AddRow(uuid.New(), uuid.New())
				// Corrected the expected SQL query.
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employee_profiles" WHERE "employee_profiles"."deleted_at" IS NULL`)).
					WillReturnRows(rows)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with no profiles",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"})
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employee_profiles" WHERE "employee_profiles"."deleted_at" IS NULL`)).
					WillReturnRows(rows)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employee_profiles" WHERE "employee_profiles"."deleted_at" IS NULL`)).
					WillReturnError(errors.New("db error"))
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			profiles, err := s.repo.GetAllEmployeeProfiles()

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, profiles, tc.wantLen)
			}
		})
	}
}
