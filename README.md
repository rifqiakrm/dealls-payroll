# Dealls Payroll Generation System

This project implements a scalable payroll generation system with predefined rules for employee attendance, overtime, and reimbursement. It is built using Go, Gin, and GORM, following SOLID principles and Clean Architecture.

## Features

* **User Management:** Employee and Admin roles with JWT-based authentication.
* **Data Seeding:** Automatically generate fake employee and admin data for development/testing.
* **Payroll Period Management:** Admin can define and manage payroll periods.
* **Employee Submissions:** Employees can submit daily attendance, overtime requests (with daily limits), and reimbursement requests.
* **Payroll Processing:** Admin can run payroll for a specific period, which calculates payslips based on attendance, overtime, and reimbursements. Once processed, records for that period are locked.
* **Payslip Generation:** Employees can generate their individual payslips with detailed breakdowns. Admin can generate a summary of all employee payslips for a period.
* **Auditing & Traceability:** Includes `created_at`, `updated_at`, `created_by`, `updated_by`, `IPAddress` for all records, and an audit log for significant changes.

## Technology Stack

* **Language:** Go
* **Web Framework:** Gin
* **ORM:** GORM
* **Database:** PostgreSQL
* **Testing:** `go.uber.org/mock` (Gomock) for unit tests, `net/http/httptest` for integration tests, `stretchr/testify` for assertions.
* **Environment Variables:** `joho/godotenv`
* **Password Hashing:** `golang.org/x/crypto/bcrypt`
* **JWT:** `dgrijalva/jwt-go`

## Getting Started

### Prerequisites

* Go (version 1.20 or higher recommended)
* PostgreSQL database server

### Installation

1. **Clone the repository:**

```bash
git clone <your-repo-url>
cd dealls-payroll
```

2. **Install Go dependencies:**

```bash
go mod tidy
```

3. **Install `mockgen` (for generating mocks for tests):**

```bash
go install go.uber.org/mock/mockgen@v0.5.2
```

4. **Generate Mocks:**

```bash
go generate ./...
```

### Configuration

Create a `.env` file in the project root (`dealls-payroll/`) with the following environment variables:

```
DB_HOST=localhost
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=deallspayslip_db
DB_PORT=5432

JWT_SECRET=your_super_secret_jwt_key

PORT=8080
GIN_MODE=release # or debug, test

ADMIN_PASSWORD=adminpassword # Optional: for the seeder
```

### Running the Application

```bash
go run cmd/server/main.go
```

The application will automatically perform database migrations on startup.

### Running the Seeder

```bash
cd db/scripts/seed
go run main.go
```

This will clear existing user-related data and create one admin user (`admin`/`ADMIN_PASSWORD` or `adminpassword`) and 100 fake employee users (`employee1` to `employee100`).

## Run with Postman

Import the Postman collection to quickly test all endpoints:

[![Run in Postman](https://run.pstmn.io/button.svg)](https://api.postman.com/collections/3441134-1d93a2ca-d617-43a9-a7fc-7905dbe2da77?access_key=PMAT-01K2WE6EDXSB5EQYBRQA2Y881H)

## API Endpoints

### Authentication

* `POST /auth/register` - Register a new user (employee or admin)
* `POST /auth/login` - Login a user and get a JWT token

### Employee Endpoints (Requires Employee JWT)

* `POST /api/employee/attendances` - Submit daily attendance
* `POST /api/employee/overtimes` - Submit overtime hours
* `POST /api/employee/reimbursements` - Submit reimbursement requests
* `POST /api/employee/payslips` - Generate individual payslip for a given payroll period
* `GET /api/employee/payroll-periods` - Get all payroll periods
* `GET /api/employee/payroll-periods/:id` - Get a payroll period by ID

### Admin Endpoints (Requires Admin JWT)

* `POST /api/admin/payroll-periods` - Create a new payroll period
* `GET /api/admin/payroll-periods` - Get all payroll periods
* `GET /api/admin/payroll-periods/:id` - Get a payroll period by ID
* `POST /api/admin/run-payroll` - Process payroll for a specific period
* `POST /api/admin/payslip-summary` - Get a summary of all payslips for a given payroll period

## Testing

To run all unit and integration tests:

```bash
go test ./... # Run all tests
```

To run tests for a specific package (e.g., services):

```bash
go test ./internal/service/...
```

## Plus Points Implementation

* **Timestamps:** `created_at` and `updated_at` are included in `domain.BaseModel` and automatically managed by GORM.
* **User Tracking:** `created_by` and `updated_by` fields are included in `domain.BaseModel` and populated in services/handlers where applicable.
* **IP Address:** `IPAddress` field is included in `domain.BaseModel` and captured from requests.
* **Audit Log:** An `AuditLog` domain model is defined. Full audit logging can be implemented using GORM hooks or service interceptors.
* **Request ID:** `request_id` is included in the `AuditLog` model for distributed tracing across services.
