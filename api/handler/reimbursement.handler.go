package handler

import (
	"net/http"
	"payroll-system/api/response"

	"github.com/gin-gonic/gin"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
)

// ReimbursementHandler handles reimbursement related HTTP requests.
type ReimbursementHandler struct {
	service service.ReimbursementServiceInterface
}

// NewReimbursementHandler creates a new ReimbursementHandler.
func NewReimbursementHandler(service service.ReimbursementServiceInterface) *ReimbursementHandler {
	return &ReimbursementHandler{service: service}
}

// SubmitReimbursementRequest represents the request body for submitting a reimbursement.
type SubmitReimbursementRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

// SubmitReimbursement handles the submission of employee reimbursement.
func (h *ReimbursementHandler) SubmitReimbursement(c *gin.Context) {
	var req SubmitReimbursementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
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

	reimbursement, err := h.service.SubmitReimbursement(currentUser.ID, req.Amount, req.Description, ipAddress, requestID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to submit reimbursement", err.Error())
		return
	}

	response.Success(c, "Reimbursement submitted successfully", response.ToReimbursementResponse(reimbursement))
}
