package handler

import (
	"net/http"
	"payroll-system/api/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
)

// PayslipHandler handles payslip related HTTP requests.
type PayslipHandler struct {
	service service.PayslipServiceInterface
}

// NewPayslipHandler creates a new PayslipHandler.
func NewPayslipHandler(service service.PayslipServiceInterface) *PayslipHandler {
	return &PayslipHandler{service: service}
}

// GetEmployeePayslipRequest represents the request for an employee to get their payslip.
type GetEmployeePayslipRequest struct {
	PayrollPeriodID string `json:"payroll_period_id" binding:"required"`
}

// GetEmployeePayslip handles an employee's request to view their payslip.
func (h *PayslipHandler) GetEmployeePayslip(c *gin.Context) {
	var req GetEmployeePayslipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	periodID, err := uuid.Parse(req.PayrollPeriodID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payroll_period_id format", nil)
		return
	}

	// Get current user
	user, exists := c.Get("currentUser")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	currentUser := user.(*domain.User)

	payslip, err := h.service.GetEmployeePayslip(currentUser.ID, periodID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve payslip", err.Error())
		return
	}

	if payslip == nil {
		response.Error(c, http.StatusNotFound, "Payslip not found", nil)
		return
	}

	response.Success(c, "Payslip retrieved successfully", response.ToPayslipResponse(payslip))
}

// GetPayslipSummary handles an admin's request to view a summary of all payslips for a period.
func (h *PayslipHandler) GetPayslipSummary(c *gin.Context) {
	var req GetEmployeePayslipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	periodID, err := uuid.Parse(req.PayrollPeriodID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payroll_period_id format", nil)
		return
	}

	payslips, totalTakeHomePay, err := h.service.GetPayslipSummaryForPeriod(periodID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve payslip summary", err.Error())
		return
	}

	// Map payslips to response DTOs
	payslipResponses := make([]response.PayslipResponse, 0)
	for _, p := range payslips {
		payslipResponses = append(payslipResponses, response.ToPayslipResponse(&p))
	}

	response.Success(c, "Payslip summary retrieved successfully", gin.H{
		"payroll_period_id":                 periodID.String(),
		"total_take_home_pay_all_employees": totalTakeHomePay,
		"payslips":                          payslipResponses,
	})
}
