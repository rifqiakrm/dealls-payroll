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

// --- Test Suite Setup for AuditLogRepository ---

type AuditLogRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo AuditLogRepository
}

// SetupSuite runs before the tests in the suite are run.
func (s *AuditLogRepositorySuite) SetupSuite() {
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
	s.repo = NewAuditLogGormRepository(db)
}

// TearDownTest runs after each test in the suite.
func (s *AuditLogRepositorySuite) TearDownTest() {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

// TestAuditLogRepository runs the test suite.
func TestAuditLogRepository(t *testing.T) {
	suite.Run(t, new(AuditLogRepositorySuite))
}

// --- Test Cases ---

func (s *AuditLogRepositorySuite) TestCreate() {
	auditID := uuid.New()
	userID := uuid.New()

	testCases := []struct {
		name    string
		audit   *domain.AuditLog
		mock    func()
		wantErr bool
	}{
		{
			name: "Success",
			audit: &domain.AuditLog{
				BaseModel: domain.BaseModel{ID: auditID},
				UserID:    &userID,
				Action:    "LOGIN",
				Timestamp: time.Now(),
			},
			mock: func() {
				s.mock.ExpectBegin()
				// This regex now precisely matches the GORM query, including the inline NULL values.
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "audit_logs" ("created_at","updated_at","deleted_at","created_by","updated_by","ip_address","user_id","action","entity_name","entity_id","old_value","new_value","request_id","timestamp","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULL,NULL,$11,$12,$13) RETURNING "id"`)).
					WithArgs(
						sqlmock.AnyArg(), // created_at
						sqlmock.AnyArg(), // updated_at
						nil,              // deleted_at
						sqlmock.AnyArg(), // created_by
						sqlmock.AnyArg(), // updated_by
						sqlmock.AnyArg(), // ip_address
						userID,           // user_id
						"LOGIN",          // action
						sqlmock.AnyArg(), // entity_name
						sqlmock.AnyArg(), // entity_id
						sqlmock.AnyArg(), // request_id
						sqlmock.AnyArg(), // timestamp
						auditID,          // id
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(auditID))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			audit: &domain.AuditLog{
				BaseModel: domain.BaseModel{ID: auditID},
				UserID:    &userID,
				Action:    "LOGIN",
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "audit_logs" ("created_at","updated_at","deleted_at","created_by","updated_by","ip_address","user_id","action","entity_name","entity_id","old_value","new_value","request_id","timestamp","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULL,NULL,$11,$12,$13) RETURNING "id"`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.Create(tc.audit)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *AuditLogRepositorySuite) TestGetByID() {
	id := uuid.New()

	testCases := []struct {
		name    string
		id      uuid.UUID
		mock    func()
		wantErr bool
		wantNil bool
	}{
		{
			name: "Success",
			id:   id,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "action"}).
					AddRow(id, "LOGIN")
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "audit_logs" WHERE id = $1 AND "audit_logs"."deleted_at" IS NULL ORDER BY "audit_logs"."id" LIMIT $2`)).
					WithArgs(id, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			id:   id,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "audit_logs" WHERE id = $1 AND "audit_logs"."deleted_at" IS NULL ORDER BY "audit_logs"."id" LIMIT $2`)).
					WithArgs(id, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			wantNil: true,
		},
		{
			name: "DB Error",
			id:   id,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "audit_logs" WHERE id = $1 AND "audit_logs"."deleted_at" IS NULL ORDER BY "audit_logs"."id" LIMIT $2`)).
					WithArgs(id, 1).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
			wantNil: false,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			audit, err := s.repo.GetByID(tc.id)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.wantNil {
				assert.Nil(t, audit)
			} else if !tc.wantErr {
				assert.NotNil(t, audit)
				assert.Equal(t, tc.id, audit.ID)
			}
		})
	}
}

func (s *AuditLogRepositorySuite) TestGetAllByUser() {
	userID := uuid.New()

	testCases := []struct {
		name    string
		limit   int
		mock    func()
		wantLen int
		wantErr bool
	}{
		{
			name:  "Success with limit",
			limit: 5,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"}).
					AddRow(uuid.New(), userID).
					AddRow(uuid.New(), userID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "audit_logs" WHERE user_id = $1 AND "audit_logs"."deleted_at" IS NULL ORDER BY timestamp desc LIMIT $2`)).
					WithArgs(userID, 5).
					WillReturnRows(rows)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:  "Success without limit",
			limit: 0,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"}).
					AddRow(uuid.New(), userID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "audit_logs" WHERE user_id = $1 AND "audit_logs"."deleted_at" IS NULL ORDER BY timestamp desc`)).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:  "DB Error",
			limit: 10,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "audit_logs" WHERE user_id = $1 AND "audit_logs"."deleted_at" IS NULL ORDER BY timestamp desc LIMIT $2`)).
					WithArgs(userID, 10).
					WillReturnError(errors.New("db error"))
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			logs, err := s.repo.GetAllByUser(userID, tc.limit)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, logs, tc.wantLen)
			}
		})
	}
}
