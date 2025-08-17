package handler

import (
	"net/http"
	"payroll-system/api/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
)

// PayrollPeriodHandler handles payroll period related HTTP requests.
type PayrollPeriodHandler struct {
	service service.PayrollPeriodServiceInterface
}

// NewPayrollPeriodHandler creates a new PayrollPeriodHandler.
func NewPayrollPeriodHandler(service service.PayrollPeriodServiceInterface) *PayrollPeriodHandler {
	return &PayrollPeriodHandler{service: service}
}

// CreatePayrollPeriodRequest represents the request body for creating a payroll period.
type CreatePayrollPeriodRequest struct {
	StartDate string `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate   string `json:"end_date" binding:"required"`   // YYYY-MM-DD
}

// CreatePayrollPeriod handles the creation of a new payroll period.
func (h *PayrollPeriodHandler) CreatePayrollPeriod(c *gin.Context) {
	var req CreatePayrollPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD.", nil)
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD.", nil)
		return
	}

	// Get current user from context (set by AuthMiddleware)
	user, exists := c.Get("currentUser")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	currentUser := user.(*domain.User)

	// Get IP address from request
	ipAddress := c.ClientIP()
	requestID := c.GetHeader("X-Request-ID")

	period, err := h.service.CreatePayrollPeriod(startDate, endDate, currentUser.ID, ipAddress, requestID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create payroll period", err.Error())
		return
	}

	response.Success(c, "Payroll period created successfully", response.ToPayrollPeriodResponse(period))
}

// GetPayrollPeriodByID handles retrieving a payroll period by its ID.
func (h *PayrollPeriodHandler) GetPayrollPeriodByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payroll period ID format", nil)
		return
	}

	period, err := h.service.GetPayrollPeriodByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve payroll period", err.Error())
		return
	}

	if period == nil {
		response.Error(c, http.StatusNotFound, "Payroll period not found", nil)
		return
	}

	response.Success(c, "Payroll period retrieved successfully", response.ToPayrollPeriodResponse(period))
}

// GetAllPayrollPeriods handles retrieving all payroll periods.
func (h *PayrollPeriodHandler) GetAllPayrollPeriods(c *gin.Context) {
	periods, err := h.service.GetAllPayrollPeriods()
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Failed to retrieve payroll periods", err.Error())
		return
	}

	data := make([]interface{}, len(periods))
	for i, v := range periods {
		data[i] = response.ToPayrollPeriodResponse(&v)
	}

	response.Success(c, "Payroll periods retrieved successfully", response.ToPayrollPeriodListResponse(periods))
}
