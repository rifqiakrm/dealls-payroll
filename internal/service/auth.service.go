package service

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"payroll-system/internal/domain"
	"payroll-system/internal/repository"
)

// AuthServiceInterface defines the methods of AuthService for mocking purposes.
//
//go:generate mockgen -source=auth.service.go -destination=../../tests/mocks/service/mock_auth_service.go -package=mocks
type AuthServiceInterface interface {
	// RegisterUser registers a new user.
	RegisterUser(username, password, role, ipAddress, requestID string) (*domain.User, error)
	// LoginUser authenticates a user and returns a JWT token.
	LoginUser(username, password, ipAddress, requestID string) (string, error)
}

// AuthService provides authentication related business logic.
type AuthService struct {
	userRepo  repository.UserRepository
	auditRepo repository.AuditLogRepository
	jwtSecret string
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, auditRepo repository.AuditLogRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		jwtSecret: jwtSecret,
	}
}

// RegisterUser registers a new user.
func (s *AuthService) RegisterUser(username, password, role string, ipAddress string, requestID string) (*domain.User, error) {
	existingUser, err := s.userRepo.GetUserByUsername(username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IPAddress: ipAddress,
		},
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	// Audit log for user creation
	_ = repository.CreateAuditLog(s.auditRepo, &user.ID, "CREATE", "User", &user.ID, nil, user, ipAddress, requestID)

	return user, nil
}

// LoginUser authenticates a user and generates a JWT token.
func (s *AuthService) LoginUser(username, password string, ipAddress string, requestID string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	// Audit log for login
	_ = repository.CreateAuditLog(s.auditRepo, &user.ID, "LOGIN", "User", &user.ID, nil, map[string]string{"ip": ipAddress}, ipAddress, requestID)

	return tokenString, nil
}
