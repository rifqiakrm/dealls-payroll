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

// --- Domain Models (for context, normally in their own package) ---

// BaseModel provides common fields for all entities.
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	CreatedBy uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	UpdatedBy uuid.UUID      `gorm:"type:uuid" json:"updated_by"`
	IPAddress string         `gorm:"type:varchar(45)" json:"ip_address"` // IPv4 or IPv6
}

// Attendance records an employee's daily attendance.
type Attendance struct {
	domain.BaseModel
	UserID          uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Date            time.Time  `gorm:"type:date;uniqueIndex:idx_user_date;not null" json:"date"`
	CheckInTime     time.Time  `gorm:"type:time;not null" json:"check_in_time"`
	CheckOutTime    time.Time  `gorm:"type:time;not null" json:"check_out_time"`
	PayrollPeriodID *uuid.UUID `gorm:"type:uuid" json:"payroll_period_id,omitempty"`
}

// --- Test Suite Setup ---

type AttendanceRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo AttendanceRepository
}

// SetupSuite runs before the tests in the suite are run.
func (s *AttendanceRepositorySuite) SetupSuite() {
	// Open a new mock database connection
	sqlDB, mock, err := sqlmock.New()
	s.Require().NoError(err)

	// Initialize GORM with the mock database
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	s.Require().NoError(err)

	s.db = db
	s.mock = mock
	s.repo = NewAttendanceGormRepository(db)
}

// TearDownTest runs after each test in the suite.
func (s *AttendanceRepositorySuite) TearDownTest() {
	// Ensure all expectations were met
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

// TestAttendanceRepository runs the test suite.
func TestAttendanceRepository(t *testing.T) {
	suite.Run(t, new(AttendanceRepositorySuite))
}

// --- Test Cases ---

func (s *AttendanceRepositorySuite) TestCreateAttendance() {
	now := time.Now()
	attendanceID := uuid.New()
	userID := uuid.New()

	testCases := []struct {
		name       string
		attendance *domain.Attendance
		mock       func()
		wantErr    bool
	}{
		{
			name: "Success",
			attendance: &domain.Attendance{
				BaseModel: domain.BaseModel{ID: attendanceID},
				UserID:    userID,
				Date:      now,
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "attendances" ("created_at","updated_at","deleted_at","created_by","updated_by","ip_address","user_id","date","check_in_time","check_out_time","payroll_period_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING "id"`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), userID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, attendanceID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(attendanceID))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "DB Error",
			attendance: &domain.Attendance{
				BaseModel: domain.BaseModel{ID: attendanceID},
				UserID:    userID,
				Date:      now,
			},
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "attendances" ("created_at","updated_at","deleted_at","created_by","updated_by","ip_address","user_id","date","check_in_time","check_out_time","payroll_period_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING "id"`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.CreateAttendance(tc.attendance)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *AttendanceRepositorySuite) TestGetAttendanceByID() {
	id := uuid.New()
	now := time.Now()

	testCases := []struct {
		name    string
		id      uuid.UUID
		mock    func()
		want    *domain.Attendance
		wantErr bool
		wantNil bool
	}{
		{
			name: "Success",
			id:   id,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "date"}).
					AddRow(id, uuid.New(), now)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE "attendances"."id" = $1 AND "attendances"."deleted_at" IS NULL ORDER BY "attendances"."id" LIMIT $2`)).
					WithArgs(id, 1).
					WillReturnRows(rows)
			},
			want:    &domain.Attendance{BaseModel: domain.BaseModel{ID: id}},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "Not Found",
			id:   id,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE "attendances"."id" = $1 AND "attendances"."deleted_at" IS NULL ORDER BY "attendances"."id" LIMIT $2`)).
					WithArgs(id, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: false,
			wantNil: true,
		},
		{
			name: "DB Error",
			id:   id,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE "attendances"."id" = $1 AND "attendances"."deleted_at" IS NULL ORDER BY "attendances"."id" LIMIT $2`)).
					WithArgs(id, 1).
					WillReturnError(errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
			wantNil: false,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			result, err := s.repo.GetAttendanceByID(tc.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, result)
			} else if tc.want != nil {
				assert.Equal(t, tc.want.ID, result.ID)
			}
		})
	}
}

func (s *AttendanceRepositorySuite) TestGetAttendanceByUserIDAndDate() {
	userID := uuid.New()
	date := time.Now()
	dateStr := date.Format("2006-01-02")

	testCases := []struct {
		name    string
		userID  uuid.UUID
		date    time.Time
		mock    func()
		want    *domain.Attendance
		wantErr bool
		wantNil bool
	}{
		{
			name:   "Success",
			userID: userID,
			date:   date,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "date"}).
					AddRow(uuid.New(), userID, date)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE (user_id = $1 AND date = $2) AND "attendances"."deleted_at" IS NULL ORDER BY "attendances"."id" LIMIT $3`)).
					WithArgs(userID, dateStr, 1).
					WillReturnRows(rows)
			},
			want:    &domain.Attendance{UserID: userID},
			wantErr: false,
			wantNil: false,
		},
		{
			name:   "Not Found",
			userID: userID,
			date:   date,
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE (user_id = $1 AND date = $2) AND "attendances"."deleted_at" IS NULL ORDER BY "attendances"."id" LIMIT $3`)).
					WithArgs(userID, dateStr, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			result, err := s.repo.GetAttendanceByUserIDAndDate(tc.userID, tc.date)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, result)
			} else if tc.want != nil {
				assert.Equal(t, tc.want.UserID, result.UserID)
			}
		})
	}
}

func (s *AttendanceRepositorySuite) TestGetAttendancesByUserIDAndPeriod() {
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
					AddRow(uuid.New(), userID).
					AddRow(uuid.New(), userID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE (user_id = $1 AND date >= $2 AND date <= $3) AND "attendances"."deleted_at" IS NULL`)).
					WithArgs(userID, startDateStr, endDateStr).
					WillReturnRows(rows)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success - No Records Found",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id"})
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE (user_id = $1 AND date >= $2 AND date <= $3) AND "attendances"."deleted_at" IS NULL`)).
					WithArgs(userID, startDateStr, endDateStr).
					WillReturnRows(rows)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE (user_id = $1 AND date >= $2 AND date <= $3) AND "attendances"."deleted_at" IS NULL`)).
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
			results, err := s.repo.GetAttendancesByUserIDAndPeriod(userID, startDate, endDate)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tc.wantLen)
			}
		})
	}
}

func (s *AttendanceRepositorySuite) TestGetAttendancesByUserIDAndPayrollPeriodID() {
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
				rows := sqlmock.NewRows([]string{"id", "user_id", "payroll_period_id"}).
					AddRow(uuid.New(), userID, payrollPeriodID).
					AddRow(uuid.New(), userID, payrollPeriodID)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE (user_id = $1 AND payroll_period_id = $2) AND "attendances"."deleted_at" IS NULL`)).
					WithArgs(userID, payrollPeriodID).
					WillReturnRows(rows)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "DB Error",
			mock: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances" WHERE (user_id = $1 AND payroll_period_id = $2) AND "attendances"."deleted_at" IS NULL`)).
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
			results, err := s.repo.GetAttendancesByUserIDAndPayrollPeriodID(userID, payrollPeriodID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tc.wantLen)
			}
		})
	}
}

func (s *AttendanceRepositorySuite) TestUpdateAttendance() {
	now := time.Now()
	att := &domain.Attendance{
		BaseModel: domain.BaseModel{ID: uuid.New(), UpdatedAt: now},
		UserID:    uuid.New(),
	}

	testCases := []struct {
		name       string
		attendance *domain.Attendance
		mock       func()
		wantErr    bool
	}{
		{
			name:       "Success",
			attendance: att,
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "attendances" SET`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), att.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:       "DB Error",
			attendance: att,
			mock: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "attendances" SET`)).
					WillReturnError(errors.New("db error"))
				s.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := s.repo.UpdateAttendance(tc.attendance)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

//func (s *AttendanceRepositorySuite) TestUpdateAttendancesTx() {
//	att1 := domain.Attendance{BaseModel: domain.BaseModel{ID: uuid.New()}}
//	att2 := domain.Attendance{BaseModel: domain.BaseModel{ID: uuid.New()}}
//	attendances := []domain.Attendance{att1, att2}
//
//	testCases := []struct {
//		name        string
//		attendances []domain.Attendance
//		mock        func()
//		wantErr     bool
//		useNilTx    bool
//	}{
//		{
//			name:        "Success",
//			attendances: attendances,
//			mock: func() {
//				// GORM's Save on an existing record is a simple UPDATE, which is an Exec.
//				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "attendances" SET`)).
//					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), att1.ID).
//					WillReturnResult(driver.ResultNoRows)
//				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "attendances" SET`)).
//					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), att2.ID).
//					WillReturnResult(driver.ResultNoRows)
//			},
//			wantErr: false,
//		},
//		{
//			name:        "DB Error on second update",
//			attendances: attendances,
//			mock: func() {
//				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "attendances" SET`)).
//					WillReturnResult(driver.ResultNoRows)
//				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "attendances" SET`)).
//					WillReturnError(errors.New("db error"))
//			},
//			wantErr: true,
//		},
//		{
//			name:     "Error with nil transaction",
//			useNilTx: true,
//			mock:     func() {},
//			wantErr:  true,
//		},
//	}
//
//	for _, tc := range testCases {
//		s.T().Run(tc.name, func(t *testing.T) {
//			if tc.useNilTx {
//				err := s.repo.UpdateAttendancesTx(nil, tc.attendances)
//				assert.Error(t, err)
//				assert.Equal(t, gorm.ErrInvalidDB, err)
//				return
//			}
//
//			s.mock.ExpectBegin()
//			tc.mock()
//			if tc.wantErr {
//				s.mock.ExpectRollback()
//			} else {
//				s.mock.ExpectCommit()
//			}
//
//			err := s.db.Transaction(func(tx *gorm.DB) error {
//				return s.repo.UpdateAttendancesTx(tx, tc.attendances)
//			})
//
//			if tc.wantErr {
//				assert.Error(t, err)
//			} else {
//				assert.NoError(t, err)
//			}
//		})
//	}
//}
