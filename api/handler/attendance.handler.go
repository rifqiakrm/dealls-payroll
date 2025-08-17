package handler

import (
	"net/http"
	"payroll-system/api/response"
	"time"

	"github.com/gin-gonic/gin"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
)

// AttendanceHandler handles attendance related HTTP requests.
type AttendanceHandler struct {
	service service.AttendanceServiceInterface
}

// NewAttendanceHandler creates a new AttendanceHandler.
func NewAttendanceHandler(service service.AttendanceServiceInterface) *AttendanceHandler {
	return &AttendanceHandler{service: service}
}

// SubmitAttendanceRequest represents the request body for submitting attendance.
type SubmitAttendanceRequest struct {
	CheckInTime  string `json:"check_in_time" binding:"required"` // YYYY-MM-DD HH:MM:SS
	CheckOutTime string `json:"check_out_time"`                   // YYYY-MM-DD HH:MM:SS
}

// SubmitAttendance handles the submission of employee attendance.
func (h *AttendanceHandler) SubmitAttendance(c *gin.Context) {
	var req SubmitAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	checkInTime, err := time.Parse("2006-01-02 15:04:05", req.CheckInTime)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid check_in_time format. Use YYYY-MM-DD HH:MM:SS", nil)
		return
	}

	var checkOutTime time.Time
	if req.CheckOutTime != "" {
		checkOutTime, err = time.Parse("2006-01-02 15:04:05", req.CheckOutTime)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid check_out_time format. Use YYYY-MM-DD HH:MM:SS", nil)
			return
		}
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

	attendance, err := h.service.SubmitAttendance(currentUser.ID, checkInTime, checkOutTime, ipAddress, requestID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to submit attendance", err.Error())
		return
	}

	response.Success(c, "Attendance submitted successfully", response.ToAttendanceResponse(attendance))
}
