package main

import (
	"log"
	"os"
	"payroll-system/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // For loading environment variables from .env file

	"payroll-system/api/handler" // Import the handler package
	"payroll-system/internal/service"

	"payroll-system/db"
	"payroll-system/internal/repository"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables.")
	}

	// Initialize database connection
	db := db.InitDB()

	// Set Gin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.ReleaseMode // Default to release mode
	}
	gin.SetMode(ginMode)

	// Initialize Gin router
	router := gin.Default()

	// --- Dependency Injection for Audit Log ---
	auditRepo := repository.NewAuditLogGormRepository(db) // GORM implementation of UserRepository

	// --- Dependency Injection for Authentication ---
	userRepo := repository.NewUserGormRepository(db) // GORM implementation of UserRepository

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set.")
	}
	authService := service.NewAuthService(userRepo, auditRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	// --- Dependency Injection for Payroll Period ---
	payrollPeriodRepo := repository.NewPayrollPeriodGormRepository(db)
	payrollPeriodService := service.NewPayrollPeriodService(payrollPeriodRepo, auditRepo)
	payrollPeriodHandler := handler.NewPayrollPeriodHandler(payrollPeriodService)

	// --- Dependency Injection for Attendance ---
	attendanceRepo := repository.NewAttendanceGormRepository(db)
	attendanceService := service.NewAttendanceService(attendanceRepo, auditRepo)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService)

	// --- Dependency Injection for Overtime ---
	overtimeRepo := repository.NewOvertimeGormRepository(db)
	overtimeService := service.NewOvertimeService(overtimeRepo, auditRepo)
	overtimeHandler := handler.NewOvertimeHandler(overtimeService)

	// --- Dependency Injection for Reimbursement ---
	reimbursementRepo := repository.NewReimbursementGormRepository(db)
	reimbursementService := service.NewReimbursementService(reimbursementRepo, auditRepo)
	reimbursementHandler := handler.NewReimbursementHandler(reimbursementService)

	// --- Dependency Injection for Employee Profile ---
	employeeProfileRepo := repository.NewEmployeeProfileGormRepository(db)

	// --- Dependency Injection for Payslip ---
	payslipRepo := repository.NewPayslipGormRepository(db)

	// --- Dependency Injection for Payroll Service ---
	payrollService := service.NewPayrollService(
		payslipRepo,
		payrollPeriodRepo,
		employeeProfileRepo,
		attendanceRepo,
		overtimeRepo,
		reimbursementRepo,
		auditRepo,
		db,
	)
	payrollHandler := handler.NewPayrollHandler(payrollService)

	// --- Dependency Injection for Payslip Service ---
	payslipService := service.NewPayslipService(payslipRepo, payrollPeriodRepo, attendanceRepo, overtimeRepo)
	payslipHandler := handler.NewPayslipHandler(payslipService)

	// --- Register API Routes ---
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	// Protected routes (example)
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(userRepo)) // Apply authentication middleware
	{
		// Example of a route that requires authentication
		protected.GET("/me", func(c *gin.Context) {
			user, _ := c.Get("currentUser")
			c.JSON(200, gin.H{"message": "Welcome!", "user": user})
		})

		// Employee-specific routes
		employeeRoutes := protected.Group("/employee")
		employeeRoutes.Use(middleware.AuthorizeMiddleware("employee")) // Apply authorization middleware for employee role
		{
			// Attendance Routes (Employee only)
			employeeRoutes.POST("/attendances", attendanceHandler.SubmitAttendance)

			// Overtime Routes (Employee only)
			employeeRoutes.POST("/overtimes", overtimeHandler.SubmitOvertime)

			// Reimbursement Routes (Employee only)
			employeeRoutes.POST("/reimbursements", reimbursementHandler.SubmitReimbursement)

			// Payslip Routes (Employee only)
			employeeRoutes.POST("/payslips", payslipHandler.GetEmployeePayslip)

			// Payroll Period Routes (Employee only)
			employeeRoutes.GET("/payroll-periods", payrollPeriodHandler.GetAllPayrollPeriods)
			employeeRoutes.GET("/payroll-periods/:id", payrollPeriodHandler.GetPayrollPeriodByID)
		}

		// Example of an admin-only route
		adminRoutes := protected.Group("/admin")
		adminRoutes.Use(middleware.AuthorizeMiddleware("admin")) // Apply authorization middleware for admin role
		{
			adminRoutes.GET("/dashboard", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Admin Dashboard"})
			})

			// Payroll Period Routes (Admin only)
			adminRoutes.POST("/payroll-periods", payrollPeriodHandler.CreatePayrollPeriod)
			adminRoutes.GET("/payroll-periods", payrollPeriodHandler.GetAllPayrollPeriods)
			adminRoutes.GET("/payroll-periods/:id", payrollPeriodHandler.GetPayrollPeriodByID)

			// Payroll Processing Routes (Admin only)
			adminRoutes.POST("/run-payroll", payrollHandler.RunPayroll)

			// Payslip Summary Routes (Admin only)
			adminRoutes.POST("/payslip-summary", payslipHandler.GetPayslipSummary)
		}
	}

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
