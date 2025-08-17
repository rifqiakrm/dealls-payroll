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

// --- Test Suite Setup for UserRepository ---

type UserRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo UserRepository
}

// SetupSuite runs before the tests in the suite are run.
func (s *UserRepositorySuite) SetupSuite() {
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
	s.repo = NewUserGormRepository(db)
}

// TearDownTest runs after each test in the suite.
func (s *UserRepositorySuite) TearDownTest() {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

// TestUserRepository runs the test suite.
func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}

// --- Test Cases ---

func (s *UserRepositorySuite) TestCreateUser() {
	userID := uuid.New()

	testCases := []struct {
		name    string
		user    *domain.User
		mock    func()
		wantErr bool
	}{
		{
			name: "Success",
			user: &domain.User{
				BaseModel: domain.BaseModel{ID: userID},
				Username:  "testuser",
				Password:  "password",
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			user: &domain.User{
				BaseModel: domain.BaseModel{ID: userID},
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.CreateUser(tc.user)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *UserRepositorySuite) TestGetUserByUsername() {
	username := "testuser"

	testCases := []struct {
		name     string
		username string
		mock     func()
		wantErr  bool
		wantNil  bool
	}{
		{
			name:     "Success",
			username: username,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "username"}).AddRow(uuid.New(), username)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
					WithArgs(username, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:     "Not Found",
			username: username,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
					WithArgs(username, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			user, err := s.repo.GetUserByUsername(tc.username)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, user)
			} else {
				assert.NotNil(t, user)
			}
		})
	}
}

func (s *UserRepositorySuite) TestGetUserByID() {
	userID := uuid.New()

	testCases := []struct {
		name    string
		id      uuid.UUID
		mock    func()
		wantErr bool
		wantNil bool
	}{
		{
			name: "Success",
			id:   userID,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(userID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
					WithArgs(userID, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			id:   userID,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
					WithArgs(userID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			user, err := s.repo.GetUserByID(tc.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, user)
			} else {
				assert.NotNil(t, user)
			}
		})
	}
}
