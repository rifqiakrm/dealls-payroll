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

func TestOvertimeHandler_SubmitOvertime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	currentUser := &domain.User{
		BaseModel: domain.BaseModel{ID: uuid.New()},
		Username:  "testuser",
	}

	dateStr := "2025-08-18"
	date, _ := time.Parse("2006-01-02", dateStr)

	testCases := []struct {
		name                 string
		requestBody          any
		setupMiddleware      func(r *gin.Engine, h *OvertimeHandler)
		mockService          func(mockService *mockSvc.MockOvertimeServiceInterface)
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "Success - Valid Overtime Submission",
			requestBody: SubmitOvertimeRequest{
				Date:  dateStr,
				Hours: 2.5,
			},
			setupMiddleware: func(r *gin.Engine, h *OvertimeHandler) {
				r.POST("/overtime", func(c *gin.Context) {
					c.Set("currentUser", currentUser)
					c.Next()
				}, h.SubmitOvertime)
			},
			mockService: func(mockService *mockSvc.MockOvertimeServiceInterface) {
				mockService.EXPECT().SubmitOvertime(currentUser.ID, date, 2.5, gomock.Any(), gomock.Any()).
					Return(&domain.Overtime{UserID: currentUser.ID, Date: date, Hours: 2.5}, nil).Times(1)
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: "Overtime submitted successfully",
		},
		{
			name:        "Error - Invalid JSON Payload",
			requestBody: `{"date": "2025-08-18", "hours": 2.5,}`,
			setupMiddleware: func(r *gin.Engine, h *OvertimeHandler) {
				r.POST("/overtime", h.SubmitOvertime)
			},
			mockService:          func(mockService *mockSvc.MockOvertimeServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - Invalid Date Format",
			requestBody: SubmitOvertimeRequest{
				Date:  "18-08-2025",
				Hours: 1,
			},
			setupMiddleware: func(r *gin.Engine, h *OvertimeHandler) {
				r.POST("/overtime", h.SubmitOvertime)
			},
			mockService:          func(mockService *mockSvc.MockOvertimeServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid date format",
		},
		{
			name: "Error - Hours Not Greater Than Zero",
			requestBody: SubmitOvertimeRequest{
				Date:  dateStr,
				Hours: 0,
			},
			setupMiddleware: func(r *gin.Engine, h *OvertimeHandler) {
				r.POST("/overtime", h.SubmitOvertime)
			},
			mockService:          func(mockService *mockSvc.MockOvertimeServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - User Not Authenticated",
			requestBody: SubmitOvertimeRequest{
				Date:  dateStr,
				Hours: 3,
			},
			setupMiddleware: func(r *gin.Engine, h *OvertimeHandler) {
				r.POST("/overtime", h.SubmitOvertime)
			},
			mockService:          func(mockService *mockSvc.MockOvertimeServiceInterface) {},
			expectedStatus:       http.StatusUnauthorized,
			expectedBodyContains: "User not authenticated",
		},
		{
			name: "Error - Service Fails to Submit",
			requestBody: SubmitOvertimeRequest{
				Date:  dateStr,
				Hours: 1.5,
			},
			setupMiddleware: func(r *gin.Engine, h *OvertimeHandler) {
				r.POST("/overtime", func(c *gin.Context) {
					c.Set("currentUser", currentUser)
					c.Next()
				}, h.SubmitOvertime)
			},
			mockService: func(mockService *mockSvc.MockOvertimeServiceInterface) {
				mockService.EXPECT().SubmitOvertime(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("service layer error")).Times(1)
			},
			expectedStatus:       http.StatusBadRequest, // Handler returns BadRequest on service error
			expectedBodyContains: "service layer error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockOvertimeService := mockSvc.NewMockOvertimeServiceInterface(ctrl)
			handler := NewOvertimeHandler(mockOvertimeService)

			tc.mockService(mockOvertimeService)

			var reqBody []byte
			if bodyStr, ok := tc.requestBody.(string); ok {
				reqBody = []byte(bodyStr)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/overtime", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			router := gin.Default()
			tc.setupMiddleware(router, handler)

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
		})
	}
}
