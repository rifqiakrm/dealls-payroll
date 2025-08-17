package handler

import (
	"net/http"
	"payroll-system/api/response"

	"github.com/gin-gonic/gin"

	"payroll-system/internal/service"
)

// AuthHandler handles authentication related HTTP requests.
type AuthHandler struct {
	authService service.AuthServiceInterface
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=employee admin"` // Enforce valid roles
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request payload",
			Data:    err.Error(),
		})
		return
	}

	// Get IP address
	ipAddress := c.ClientIP()
	requestID := c.GetHeader("X-Request-ID")

	user, err := h.authService.RegisterUser(req.Username, req.Password, req.Role, ipAddress, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to register user",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.APIResponse{
		Code:    http.StatusCreated,
		Message: "User registered successfully",
		Data: gin.H{
			"user_id":  user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login and returns a JWT token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request payload",
			Data:    err.Error(),
		})
		return
	}

	// Get IP address
	ipAddress := c.ClientIP()
	requestID := c.GetHeader("X-Request-ID")

	token, err := h.authService.LoginUser(req.Username, req.Password, ipAddress, requestID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.APIResponse{
			Code:    http.StatusUnauthorized,
			Message: "Invalid username or password",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Code:    http.StatusOK,
		Message: "Login successful",
		Data: gin.H{
			"token": token,
		},
	})
}
