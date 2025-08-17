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

func TestReimbursementHandler_SubmitReimbursement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	currentUser := &domain.User{
		BaseModel: domain.BaseModel{ID: uuid.New()},
		Username:  "testuser",
	}

	testCases := []struct {
		name                 string
		requestBody          any
		setupMiddleware      func(r *gin.Engine, h *ReimbursementHandler)
		mockService          func(mockService *mockSvc.MockReimbursementServiceInterface)
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "Success - Valid Submission",
			requestBody: SubmitReimbursementRequest{
				Amount:      150.75,
				Description: "Team Lunch",
			},
			setupMiddleware: func(r *gin.Engine, h *ReimbursementHandler) {
				r.POST("/reimbursements", func(c *gin.Context) { c.Set("currentUser", currentUser); c.Next() }, h.SubmitReimbursement)
			},
			mockService: func(mockService *mockSvc.MockReimbursementServiceInterface) {
				mockService.EXPECT().SubmitReimbursement(currentUser.ID, 150.75, "Team Lunch", gomock.Any(), gomock.Any()).
					Return(&domain.Reimbursement{UserID: currentUser.ID, Amount: 150.75}, nil).Times(1)
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: "Reimbursement submitted successfully",
		},
		{
			name:        "Error - Invalid JSON",
			requestBody: `{"amount": 100,,}`,
			setupMiddleware: func(r *gin.Engine, h *ReimbursementHandler) {
				r.POST("/reimbursements", h.SubmitReimbursement)
			},
			mockService:          func(mockService *mockSvc.MockReimbursementServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - Amount is Zero",
			requestBody: SubmitReimbursementRequest{
				Amount: 0,
			},
			setupMiddleware: func(r *gin.Engine, h *ReimbursementHandler) {
				r.POST("/reimbursements", h.SubmitReimbursement)
			},
			mockService:          func(mockService *mockSvc.MockReimbursementServiceInterface) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid request payload",
		},
		{
			name: "Error - User Not Authenticated",
			requestBody: SubmitReimbursementRequest{
				Amount: 100,
			},
			setupMiddleware: func(r *gin.Engine, h *ReimbursementHandler) {
				r.POST("/reimbursements", h.SubmitReimbursement)
			},
			mockService:          func(mockService *mockSvc.MockReimbursementServiceInterface) {},
			expectedStatus:       http.StatusUnauthorized,
			expectedBodyContains: "User not authenticated",
		},
		{
			name: "Error - Service Failure",
			requestBody: SubmitReimbursementRequest{
				Amount: 50.0,
			},
			setupMiddleware: func(r *gin.Engine, h *ReimbursementHandler) {
				r.POST("/reimbursements", func(c *gin.Context) { c.Set("currentUser", currentUser); c.Next() }, h.SubmitReimbursement)
			},
			mockService: func(mockService *mockSvc.MockReimbursementServiceInterface) {
				mockService.EXPECT().SubmitReimbursement(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("service layer error")).Times(1)
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedBodyContains: "Failed to submit reimbursement",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockService := mockSvc.NewMockReimbursementServiceInterface(ctrl)
			handler := NewReimbursementHandler(mockService)

			tc.mockService(mockService)

			var reqBody []byte
			if bodyStr, ok := tc.requestBody.(string); ok {
				reqBody = []byte(bodyStr)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/reimbursements", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			router := gin.Default()
			tc.setupMiddleware(router, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
		})
	}
}
