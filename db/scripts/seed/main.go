package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"payroll-system/internal/domain"
	"payroll-system/internal/infrastructure/database"
)

func main() {
	// Load environment variables
	err := godotenv.Load("../../.env") // Adjust path if .env is in root
	if err != nil {
		log.Println("No .env file found, relying on environment variables.")
	}

	db := database.InitDB() // Initialize DB connection and run migrations

	// Clear existing data (optional, for fresh seeding)
	log.Println("Clearing existing data...")
	db.Exec("DELETE FROM audit_logs")
	db.Exec("DELETE FROM payslips")
	db.Exec("DELETE FROM reimbursements")
	db.Exec("DELETE FROM overtimes")
	db.Exec("DELETE FROM attendances")
	db.Exec("DELETE FROM employee_profiles")
	db.Exec("DELETE FROM payroll_periods")
	db.Exec("DELETE FROM users")
	log.Println("Existing data cleared.")

	// Seed Admin User
	log.Println("Seeding admin user...")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "adminpassword" // Default admin password
		log.Printf("ADMIN_PASSWORD not set, using default: %s", adminPassword)
	}
	hashedAdminPassword, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	adminUser := &domain.User{
		Username: "admin",
		Password: string(hashedAdminPassword),
		Role:     "admin",
	}
	if err := db.Create(adminUser).Error; err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}
	log.Println("Admin user seeded.")

	// Seed 100 Fake Employees
	log.Println("Seeding 100 fake employees...")
	for i := 1; i <= 100; i++ {
		username := fmt.Sprintf("employee%d", i)
		password := fmt.Sprintf("password%d", i)
		salary := float64(5000000 + i*10000) // Example salary range

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		employeeUser := &domain.User{
			Username: username,
			Password: string(hashedPassword),
			Role:     "employee",
		}
		if err := db.Create(employeeUser).Error; err != nil {
			log.Fatalf("Failed to seed employee %d: %v", i, err)
		}

		employeeProfile := &domain.EmployeeProfile{
			UserID: employeeUser.ID,
			Salary: salary,
		}
		if err := db.Create(employeeProfile).Error; err != nil {
			log.Fatalf("Failed to seed employee profile %d: %v", i, err)
		}
	}
	log.Println("100 fake employees seeded successfully.")

	log.Println("Database seeding completed!")
}
