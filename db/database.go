package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
)

// InitDB initializes the database connection and performs auto-migrations.
func InitDB() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&domain.User{},
		&domain.EmployeeProfile{},
		&domain.PayrollPeriod{},
		&domain.Attendance{},
		&domain.Overtime{},
		&domain.Reimbursement{},
		&domain.Payslip{},
		&domain.AuditLog{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database schema: %v", err)
	}

	log.Println("Database connection established and schema migrated successfully.")
	return db
}
