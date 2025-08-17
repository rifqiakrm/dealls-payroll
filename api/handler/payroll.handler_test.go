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

func TestPayrollHandler_RunPayroll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	currentUser := &domain.User{
		BaseModel: domain.BaseModel{ID: uuid.New()},
		Username:  "adminuser",
		Role:      "admin",
	}
	periodID := uuid.New()

	testCases := []struct {
		name                 string
		requestBody          any
		setupMiddleware      func(r *gin.Engine, h *PayrollHandler)
		mockService          func(mockService *mockSvc.MockPayrollServiceInterface)
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "Success - Payroll Run Successfully",
			requestBody: RunPayrollRequest{
				PayrollPeriodID: periodID.String(),
			},
			setupMiddleware: func(r *gin.Engine, h *PayrollHandler) {
				r.POST("/payroll/run", func(c *gin.Context) {
					c.Set("currentUser", currentUser)
					c.Next()
				}, h.RunPayroll)
			},
			mockService: func(mockService *mockSvc.MockPayrollServiceInterface) {
				mockService.EXPECT().RunPayroll(periodID, currentUser.ID, gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: "Payroll processed successfully",
		},
		{
			name:        "Error - Invalid JSON Payload",
			requestBody: `{"payroll_period_id": "invalid}`,
			setupMiddleware: func(r *gin.Engine, h *PayrollHandler) {
				r.POST("/payroll/run", h.RunPayroll)
			},
			mockService:          func(mockService *mockSvc.MockPayrollServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - Invalid Payroll Period ID Format",
			requestBody: RunPayrollRequest{
				PayrollPeriodID: "not-a-uuid",
			},
			setupMiddleware: func(r *gin.Engine, h *PayrollHandler) {
				r.POST("/payroll/run", h.RunPayroll)
			},
			mockService:          func(mockService *mockSvc.MockPayrollServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid payroll_period_id format",
		},
		{
			name: "Error - User Not Authenticated",
			requestBody: RunPayrollRequest{
				PayrollPeriodID: periodID.String(),
			},
			setupMiddleware: func(r *gin.Engine, h *PayrollHandler) {
				r.POST("/payroll/run", h.RunPayroll)
			},
			mockService:          func(mockService *mockSvc.MockPayrollServiceInterface) {},
			expectedStatus:       http.StatusUnauthorized,
			expectedBodyContains: "User not authenticated",
		},
		{
			name: "Error - Service Fails to Process Payroll",
			requestBody: RunPayrollRequest{
				PayrollPeriodID: periodID.String(),
			},
			setupMiddleware: func(r *gin.Engine, h *PayrollHandler) {
				r.POST("/payroll/run", func(c *gin.Context) {
					c.Set("currentUser", currentUser)
					c.Next()
				}, h.RunPayroll)
			},
			mockService: func(mockService *mockSvc.MockPayrollServiceInterface) {
				mockService.EXPECT().RunPayroll(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("service layer error")).Times(1)
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedBodyContains: "Failed to process payroll",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockPayrollService := mockSvc.NewMockPayrollServiceInterface(ctrl)
			handler := NewPayrollHandler(mockPayrollService)

			tc.mockService(mockPayrollService)

			var reqBody []byte
			if bodyStr, ok := tc.requestBody.(string); ok {
				reqBody = []byte(bodyStr)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/payroll/run", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			router := gin.Default()
			tc.setupMiddleware(router, handler)

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
		})
	}
}
