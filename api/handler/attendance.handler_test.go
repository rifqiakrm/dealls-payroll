package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"payroll-system/internal/domain"
	mockSvc "payroll-system/tests/mocks/service"
)

// TestSubmitAttendance provides test coverage for the SubmitAttendance handler function.
func TestSubmitAttendance(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Define a standard user for authenticated requests.
	currentUser := &domain.User{
		BaseModel: domain.BaseModel{ID: uuid.New()},
		Username:  "testuser",
	}

	// Define standard check-in and check-out times for valid requests.
	checkInStr := "2025-08-18 09:00:00"
	checkOutStr := "2025-08-18 17:00:00"
	checkInTime, _ := time.Parse("2006-01-02 15:04:05", checkInStr)
	checkOutTime, _ := time.Parse("2006-01-02 15:04:05", checkOutStr)

	testCases := []struct {
		name                 string
		requestBody          any
		setupMiddleware      func(r *gin.Engine, h *AttendanceHandler)
		mockService          func(mockService *mockSvc.MockAttendanceServiceInterface)
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "Success - Full Attendance",
			requestBody: SubmitAttendanceRequest{
				CheckInTime:  checkInStr,
				CheckOutTime: checkOutStr,
			},
			setupMiddleware: func(r *gin.Engine, h *AttendanceHandler) {
				r.POST("/attendance", func(c *gin.Context) {
					c.Set("currentUser", currentUser)
					c.Next()
				}, h.SubmitAttendance)
			},
			mockService: func(mockService *mockSvc.MockAttendanceServiceInterface) {
				mockService.EXPECT().SubmitAttendance(currentUser.ID, checkInTime, checkOutTime, gomock.Any(), gomock.Any()).
					Return(&domain.Attendance{UserID: currentUser.ID}, nil).Times(1)
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: "Attendance submitted successfully",
		},
		{
			name: "Success - Check-in Only",
			requestBody: SubmitAttendanceRequest{
				CheckInTime: checkInStr,
			},
			setupMiddleware: func(r *gin.Engine, h *AttendanceHandler) {
				r.POST("/attendance", func(c *gin.Context) {
					c.Set("currentUser", currentUser)
					c.Next()
				}, h.SubmitAttendance)
			},
			mockService: func(mockService *mockSvc.MockAttendanceServiceInterface) {
				mockService.EXPECT().SubmitAttendance(currentUser.ID, checkInTime, time.Time{}, gomock.Any(), gomock.Any()).
					Return(&domain.Attendance{UserID: currentUser.ID}, nil).Times(1)
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: "Attendance submitted successfully",
		},
		{
			name:        "Error - Invalid JSON Payload",
			requestBody: `{"check_in_time": "2025-08-18 09:00:00",,}`, // Malformed JSON
			setupMiddleware: func(r *gin.Engine, h *AttendanceHandler) {
				r.POST("/attendance", h.SubmitAttendance)
			},
			mockService:          func(mockService *mockSvc.MockAttendanceServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - Invalid Check-in Time Format",
			requestBody: SubmitAttendanceRequest{
				CheckInTime: "18-08-2025 09:00", // Wrong format
			},
			setupMiddleware: func(r *gin.Engine, h *AttendanceHandler) {
				r.POST("/attendance", h.SubmitAttendance)
			},
			mockService:          func(mockService *mockSvc.MockAttendanceServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid check_in_time format",
		},
		{
			name: "Error - Invalid Check-out Time Format",
			requestBody: SubmitAttendanceRequest{
				CheckInTime:  checkInStr,
				CheckOutTime: "invalid-date",
			},
			setupMiddleware: func(r *gin.Engine, h *AttendanceHandler) {
				r.POST("/attendance", h.SubmitAttendance)
			},
			mockService:          func(mockService *mockSvc.MockAttendanceServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid check_out_time format",
		},
		{
			name: "Error - User Not Authenticated",
			requestBody: SubmitAttendanceRequest{
				CheckInTime: checkInStr,
			},
			setupMiddleware: func(r *gin.Engine, h *AttendanceHandler) {
				// No middleware to set the user
				r.POST("/attendance", h.SubmitAttendance)
			},
			mockService:          func(mockService *mockSvc.MockAttendanceServiceInterface) {},
			expectedStatus:       http.StatusUnauthorized,
			expectedBodyContains: "User not authenticated",
		},
		{
			name: "Error - Service Fails to Submit",
			requestBody: SubmitAttendanceRequest{
				CheckInTime: checkInStr,
			},
			setupMiddleware: func(r *gin.Engine, h *AttendanceHandler) {
				r.POST("/attendance", func(c *gin.Context) {
					c.Set("currentUser", currentUser)
					c.Next()
				}, h.SubmitAttendance)
			},
			mockService: func(mockService *mockSvc.MockAttendanceServiceInterface) {
				mockService.EXPECT().SubmitAttendance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("service layer error")).Times(1)
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedBodyContains: "Failed to submit attendance",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup gomock controller and mock service
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockService := mockSvc.NewMockAttendanceServiceInterface(ctrl)
			handler := NewAttendanceHandler(mockService)

			// Set up mock expectations for this specific test case.
			tc.mockService(mockService)

			// Marshal the request body.
			var reqBody []byte
			if bodyStr, ok := tc.requestBody.(string); ok {
				reqBody = []byte(bodyStr)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			// Create the HTTP request and response recorder.
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/attendance", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a new Gin engine for each test case.
			router := gin.Default()
			tc.setupMiddleware(router, handler)

			// Serve the request.
			router.ServeHTTP(w, req)

			// Assert the results.
			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
		})
	}
}
