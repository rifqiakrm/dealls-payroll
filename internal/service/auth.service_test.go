package service_test

import (
	"errors"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"payroll-system/internal/domain"
	"payroll-system/internal/service"
	mockRepo "payroll-system/tests/mocks/repository"
)

func TestAuthService_RegisterUser(t *testing.T) {
	username := "johndoe"
	password := "password123"
	role := "admin"
	ip := "127.0.0.1"
	requestID := "req-1"

	tests := []struct {
		name            string
		mockExisting    *domain.User
		mockGetError    error
		mockCreateError error
		expectedError   string
	}{
		{
			name:          "username already exists",
			mockExisting:  &domain.User{BaseModel: domain.BaseModel{ID: uuid.New()}, Username: username},
			expectedError: "user with this username already exists",
		},
		{
			name:          "repo get error",
			mockGetError:  errors.New("db error"),
			expectedError: "db error",
		},
		{
			name:            "create user error",
			mockExisting:    nil,
			mockCreateError: errors.New("create failed"),
			expectedError:   "create failed",
		},
		{
			name:         "successful registration",
			mockExisting: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mockRepo.NewMockUserRepository(ctrl)
			mockAuditRepo := mockRepo.NewMockAuditLogRepository(ctrl)
			svc := service.NewAuthService(mockUserRepo, mockAuditRepo, "secret")

			mockUserRepo.EXPECT().
				GetUserByUsername(username).
				Return(tt.mockExisting, tt.mockGetError).
				AnyTimes()

			if tt.mockExisting == nil {
				mockUserRepo.EXPECT().
					CreateUser(gomock.Any()).
					Return(tt.mockCreateError).
					AnyTimes()
				mockAuditRepo.EXPECT().
					Create(gomock.Any()).
					Return(nil).
					AnyTimes()
			}

			user, err := svc.RegisterUser(username, password, role, ip, requestID)

			if tt.expectedError != "" {
				assert.Nil(t, user)
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NotNil(t, user)
				assert.NoError(t, err)
				assert.Equal(t, username, user.Username)
				assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)))
			}
		})
	}
}

func TestAuthService_LoginUser(t *testing.T) {
	username := "johndoe"
	password := "password123"
	ip := "127.0.0.1"
	requestID := "req-1"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	userID := uuid.New()

	tests := []struct {
		name         string
		mockUser     *domain.User
		mockGetError error
		inputPass    string
		expectedErr  string
	}{
		{
			name:         "repo get error",
			mockUser:     nil,
			mockGetError: errors.New("db error"),
			inputPass:    password,
			expectedErr:  "db error",
		},
		{
			name:        "user not found",
			mockUser:    nil,
			inputPass:   password,
			expectedErr: "invalid credentials",
		},
		{
			name:        "wrong password",
			mockUser:    &domain.User{BaseModel: domain.BaseModel{ID: userID}, Username: username, Password: string(hashedPassword)},
			inputPass:   "wrongpass",
			expectedErr: "invalid credentials",
		},
		{
			name:      "successful login",
			mockUser:  &domain.User{BaseModel: domain.BaseModel{ID: userID}, Username: username, Password: string(hashedPassword), Role: "admin"},
			inputPass: password,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mockRepo.NewMockUserRepository(ctrl)
			mockAuditRepo := mockRepo.NewMockAuditLogRepository(ctrl)
			svc := service.NewAuthService(mockUserRepo, mockAuditRepo, "secret")

			mockUserRepo.EXPECT().
				GetUserByUsername(username).
				Return(tt.mockUser, tt.mockGetError).
				AnyTimes()

			if tt.mockUser != nil && tt.inputPass == password {
				mockAuditRepo.EXPECT().
					Create(gomock.Any()).
					Return(nil).
					AnyTimes()
			}

			token, err := svc.LoginUser(username, tt.inputPass, ip, requestID)

			if tt.expectedErr != "" {
				assert.Empty(t, token)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NotEmpty(t, token)
				assert.NoError(t, err)

				parsedToken, parseErr := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte("secret"), nil
				})
				assert.NoError(t, parseErr)
				claims := parsedToken.Claims.(jwt.MapClaims)
				assert.Equal(t, username, claims["username"])
			}
		})
	}
}
