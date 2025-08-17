package handler

import (
	"net/http"
	"payroll-system/api/response"
	"time"

	"github.com/gin-gonic/gin"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
)

// OvertimeHandler handles overtime related HTTP requests.
type OvertimeHandler struct {
	service service.OvertimeServiceInterface
}

// NewOvertimeHandler creates a new OvertimeHandler.
func NewOvertimeHandler(service service.OvertimeServiceInterface) *OvertimeHandler {
	return &OvertimeHandler{service: service}
}

// SubmitOvertimeRequest represents the request body for submitting overtime.
type SubmitOvertimeRequest struct {
	Date  string  `json:"date" binding:"required"` // YYYY-MM-DD
	Hours float64 `json:"hours" binding:"required,gt=0"`
}

// SubmitOvertime handles the submission of employee overtime.
func (h *OvertimeHandler) SubmitOvertime(c *gin.Context) {
	var req SubmitOvertimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", nil)
		return
	}

	// Get current user from context
	user, exists := c.Get("currentUser")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	currentUser := user.(*domain.User)

	// Get IP address
	ipAddress := c.ClientIP()
	requestID := c.GetHeader("X-Request-ID")

	overtime, err := h.service.SubmitOvertime(currentUser.ID, date, req.Hours, ipAddress, requestID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(c, "Overtime submitted successfully", response.ToOvertimeResponse(overtime))
}
