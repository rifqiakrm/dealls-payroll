package handler

import (
	"net/http"
	"payroll-system/api/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
)

// PayrollHandler handles payroll related HTTP requests.
type PayrollHandler struct {
	service service.PayrollServiceInterface
}

// NewPayrollHandler creates a new PayrollHandler.
func NewPayrollHandler(service service.PayrollServiceInterface) *PayrollHandler {
	return &PayrollHandler{service: service}
}

// RunPayrollRequest represents the request body for running payroll.
type RunPayrollRequest struct {
	PayrollPeriodID string `json:"payroll_period_id" binding:"required"`
}

// RunPayroll handles the request to process payroll for a given period.
func (h *PayrollHandler) RunPayroll(c *gin.Context) {
	var req RunPayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	periodID, err := uuid.Parse(req.PayrollPeriodID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payroll_period_id format", nil)
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

	if err := h.service.RunPayroll(periodID, currentUser.ID, ipAddress, requestID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to process payroll", err.Error())
		return
	}

	response.Success(c, "Payroll processed successfully", nil)
}
