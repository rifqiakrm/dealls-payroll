package response

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard structure for all API responses
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success returns a success response with single object
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}

// Error returns a standardized error response
func Error(c *gin.Context, code int, message string, data interface{}) {
	if os.Getenv("GIN_MODE") == "release" {
		data = nil
	}

	c.JSON(code, APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
