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

func TestPayslipHandler_GetEmployeePayslip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	currentUser := &domain.User{
		BaseModel: domain.BaseModel{ID: uuid.New()},
		Username:  "testuser",
	}
	periodID := uuid.New()

	testCases := []struct {
		name                 string
		requestBody          any
		setupMiddleware      func(r *gin.Engine, h *PayslipHandler)
		mockService          func(mockService *mockSvc.MockPayslipServiceInterface)
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "Success - Get Employee Payslip",
			requestBody: GetEmployeePayslipRequest{
				PayrollPeriodID: periodID.String(),
			},
			setupMiddleware: func(r *gin.Engine, h *PayslipHandler) {
				r.POST("/payslip", func(c *gin.Context) { c.Set("currentUser", currentUser); c.Next() }, h.GetEmployeePayslip)
			},
			mockService: func(mockService *mockSvc.MockPayslipServiceInterface) {
				mockService.EXPECT().GetEmployeePayslip(currentUser.ID, periodID).
					Return(&domain.Payslip{UserID: currentUser.ID}, nil).Times(1)
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: "Payslip retrieved successfully",
		},
		{
			name:        "Error - Invalid JSON",
			requestBody: `{"payroll_period_id": "invalid}`,
			setupMiddleware: func(r *gin.Engine, h *PayslipHandler) {
				r.POST("/payslip", h.GetEmployeePayslip)
			},
			mockService:          func(mockService *mockSvc.MockPayslipServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - Invalid Period ID Format",
			requestBody: GetEmployeePayslipRequest{
				PayrollPeriodID: "not-a-uuid",
			},
			setupMiddleware: func(r *gin.Engine, h *PayslipHandler) {
				r.POST("/payslip", h.GetEmployeePayslip)
			},
			mockService:          func(mockService *mockSvc.MockPayslipServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid payroll_period_id format",
		},
		{
			name: "Error - User Not Authenticated",
			requestBody: GetEmployeePayslipRequest{
				PayrollPeriodID: periodID.String(),
			},
			setupMiddleware: func(r *gin.Engine, h *PayslipHandler) {
				r.POST("/payslip", h.GetEmployeePayslip)
			},
			mockService:          func(mockService *mockSvc.MockPayslipServiceInterface) {},
			expectedStatus:       http.StatusUnauthorized,
			expectedBodyContains: "User not authenticated",
		},
		{
			name: "Error - Payslip Not Found",
			requestBody: GetEmployeePayslipRequest{
				PayrollPeriodID: periodID.String(),
			},
			setupMiddleware: func(r *gin.Engine, h *PayslipHandler) {
				r.POST("/payslip", func(c *gin.Context) { c.Set("currentUser", currentUser); c.Next() }, h.GetEmployeePayslip)
			},
			mockService: func(mockService *mockSvc.MockPayslipServiceInterface) {
				mockService.EXPECT().GetEmployeePayslip(currentUser.ID, periodID).Return(nil, nil).Times(1)
			},
			expectedStatus:       http.StatusNotFound,
			expectedBodyContains: "Payslip not found",
		},
		{
			name: "Error - Service Failure",
			requestBody: GetEmployeePayslipRequest{
				PayrollPeriodID: periodID.String(),
			},
			setupMiddleware: func(r *gin.Engine, h *PayslipHandler) {
				r.POST("/payslip", func(c *gin.Context) { c.Set("currentUser", currentUser); c.Next() }, h.GetEmployeePayslip)
			},
			mockService: func(mockService *mockSvc.MockPayslipServiceInterface) {
				mockService.EXPECT().GetEmployeePayslip(currentUser.ID, periodID).Return(nil, errors.New("db error")).Times(1)
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedBodyContains: "Failed to retrieve payslip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockService := mockSvc.NewMockPayslipServiceInterface(ctrl)
			handler := NewPayslipHandler(mockService)

			tc.mockService(mockService)

			var reqBody []byte
			if bodyStr, ok := tc.requestBody.(string); ok {
				reqBody = []byte(bodyStr)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/payslip", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			router := gin.Default()
			tc.setupMiddleware(router, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
		})
	}
}

func TestPayslipHandler_GetPayslipSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	periodID := uuid.New()

	testCases := []struct {
		name                 string
		requestBody          any
		mockService          func(mockService *mockSvc.MockPayslipServiceInterface)
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "Success - Get Summary",
			requestBody: GetEmployeePayslipRequest{
				PayrollPeriodID: periodID.String(),
			},
			mockService: func(mockService *mockSvc.MockPayslipServiceInterface) {
				payslips := []domain.Payslip{{}, {}}
				mockService.EXPECT().GetPayslipSummaryForPeriod(periodID).Return(payslips, 150000.0, nil).Times(1)
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: "Payslip summary retrieved successfully",
		},
		{
			name:                 "Error - Invalid JSON",
			requestBody:          `{"payroll_period_id": "invalid}`,
			mockService:          func(mockService *mockSvc.MockPayslipServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - Invalid Period ID",
			requestBody: GetEmployeePayslipRequest{
				PayrollPeriodID: "not-a-uuid",
			},
			mockService:          func(mockService *mockSvc.MockPayslipServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid payroll_period_id format",
		},
		{
			name: "Error - Service Failure",
			requestBody: GetEmployeePayslipRequest{
				PayrollPeriodID: periodID.String(),
			},
			mockService: func(mockService *mockSvc.MockPayslipServiceInterface) {
				mockService.EXPECT().GetPayslipSummaryForPeriod(periodID).Return(nil, 0.0, errors.New("db error")).Times(1)
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedBodyContains: "Failed to retrieve payslip summary",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockService := mockSvc.NewMockPayslipServiceInterface(ctrl)
			handler := NewPayslipHandler(mockService)

			tc.mockService(mockService)

			var reqBody []byte
			if bodyStr, ok := tc.requestBody.(string); ok {
				reqBody = []byte(bodyStr)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/summary", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			router := gin.Default()
			router.POST("/summary", handler.GetPayslipSummary)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
		})
	}
}
