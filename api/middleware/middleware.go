package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"payroll-system/internal/domain"
	"payroll-system/internal/repository"
)

// AuthMiddleware authenticates requests using JWT.
func AuthMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT_SECRET not configured"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what we expect: HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token claims"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		user, err := userRepo.GetUserByID(userID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("currentUser", user)
		c.Next()
	}
}

// AuthorizeMiddleware checks if the current user has one of the required roles.
func AuthorizeMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		currentUser, ok := user.(*domain.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
			c.Abort()
			return
		}

		for _, role := range roles {
			if currentUser.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}
