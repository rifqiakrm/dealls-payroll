package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"payroll-system/internal/domain"
	mockSvc "payroll-system/tests/mocks/service"
)

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name                 string
		requestBody          any
		mockService          func(mockService *mockSvc.MockAuthServiceInterface)
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "Success - Register Employee",
			requestBody: RegisterRequest{
				Username: "newuser",
				Password: "password123",
				Role:     "employee",
			},
			mockService: func(mockService *mockSvc.MockAuthServiceInterface) {
				mockService.EXPECT().RegisterUser("newuser", "password123", "employee", gomock.Any(), gomock.Any()).
					Return(&domain.User{
						BaseModel: domain.BaseModel{ID: uuid.New()},
						Username:  "newuser",
						Role:      "employee",
					}, nil).Times(1)
			},
			expectedStatus:       http.StatusCreated,
			expectedBodyContains: "User registered successfully",
		},
		{
			name:                 "Error - Invalid JSON Payload",
			requestBody:          `{"username": "badjson",}`,
			mockService:          func(mockService *mockSvc.MockAuthServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - Invalid Role",
			requestBody: RegisterRequest{
				Username: "test",
				Password: "password",
				Role:     "guest", // Invalid role
			},
			mockService:          func(mockService *mockSvc.MockAuthServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - Service Fails to Register",
			requestBody: RegisterRequest{
				Username: "existinguser",
				Password: "password123",
				Role:     "admin",
			},
			mockService: func(mockService *mockSvc.MockAuthServiceInterface) {
				mockService.EXPECT().RegisterUser("existinguser", "password123", "admin", gomock.Any(), gomock.Any()).
					Return(nil, errors.New("username already exists")).Times(1)
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedBodyContains: "Failed to register user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockAuthService := mockSvc.NewMockAuthServiceInterface(ctrl)
			handler := NewAuthHandler(mockAuthService)

			tc.mockService(mockAuthService)

			var reqBody []byte
			if bodyStr, ok := tc.requestBody.(string); ok {
				reqBody = []byte(bodyStr)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			router := gin.Default()
			router.POST("/register", handler.Register)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name                 string
		requestBody          any
		mockService          func(mockService *mockSvc.MockAuthServiceInterface)
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "Success - Valid Login",
			requestBody: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockService: func(mockService *mockSvc.MockAuthServiceInterface) {
				mockService.EXPECT().LoginUser("testuser", "password123", gomock.Any(), gomock.Any()).
					Return("some.jwt.token", nil).Times(1)
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: "Login successful",
		},
		{
			name: "Error - Invalid Credentials",
			requestBody: LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockService: func(mockService *mockSvc.MockAuthServiceInterface) {
				mockService.EXPECT().LoginUser("testuser", "wrongpassword", gomock.Any(), gomock.Any()).
					Return("", errors.New("invalid credentials")).Times(1)
			},
			expectedStatus:       http.StatusUnauthorized,
			expectedBodyContains: "Invalid username or password",
		},
		{
			name: "Error - Invalid JSON Payload",
			requestBody: LoginRequest{
				Username: "testuser", // Missing password
			},
			mockService:          func(mockService *mockSvc.MockAuthServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockAuthService := mockSvc.NewMockAuthServiceInterface(ctrl)
			handler := NewAuthHandler(mockAuthService)

			tc.mockService(mockAuthService)

			reqBody, _ := json.Marshal(tc.requestBody)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			router := gin.Default()
			router.POST("/login", handler.Login)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
		})
	}
}
