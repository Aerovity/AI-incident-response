package services

import (
	"fmt"
	"incident-backend/internal/models"
	"incident-backend/internal/storage"
	"incident-backend/internal/utils"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT claims
type Claims struct {
	UserID         string          `json:"user_id"`
	OrganizationID string          `json:"organization_id"`
	Role           models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// AuthService handles authentication logic
type AuthService struct {
	userStore *storage.UserStore
	orgStore  *storage.OrgStore
	jwtSecret string
	jwtExpiry time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(userStore *storage.UserStore, orgStore *storage.OrgStore, jwtSecret string, jwtExpiry time.Duration) *AuthService {
	return &AuthService{
		userStore: userStore,
		orgStore:  orgStore,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email            string `json:"email"`
	Name             string `json:"name"`
	Password         string `json:"password"`
	OrganizationName string `json:"organization_name"`
	OrganizationSlug string `json:"organization_slug"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	OrganizationSlug string `json:"organization_slug"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token        string               `json:"token"`
	User         models.UserResponse  `json:"user"`
	Organization models.Organization  `json:"organization"`
}

// Register creates a new user and organization
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validate inputs
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired(req.Name, "name"); err != nil {
		return nil, err
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired(req.OrganizationName, "organization_name"); err != nil {
		return nil, err
	}
	if err := utils.ValidateSlug(req.OrganizationSlug); err != nil {
		return nil, err
	}

	// Check if organization slug already exists
	_, err := s.orgStore.GetBySlug(req.OrganizationSlug)
	if err == nil {
		return nil, fmt.Errorf("organization with slug '%s' already exists", req.OrganizationSlug)
	}

	// Create organization
	now := time.Now()
	org := &models.Organization{
		ID:              uuid.New().String(),
		Name:            req.OrganizationName,
		Slug:            strings.ToLower(req.OrganizationSlug),
		Plan:            "free",
		CreatedAt:       now,
		UpdatedAt:       now,
		AutoRemediation: false,
	}

	if err := s.orgStore.Create(org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user (first user is admin)
	user := &models.User{
		ID:             uuid.New().String(),
		Email:          strings.ToLower(strings.TrimSpace(req.Email)),
		Name:           req.Name,
		PasswordHash:   hashedPassword,
		Role:           models.RoleAdmin, // First user is always admin
		OrganizationID: org.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.userStore.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token:        token,
		User:         user.ToResponse(),
		Organization: *org,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Validate inputs
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired(req.Password, "password"); err != nil {
		return nil, err
	}
	if err := utils.ValidateSlug(req.OrganizationSlug); err != nil {
		return nil, err
	}

	// Get organization
	org, err := s.orgStore.GetBySlug(req.OrganizationSlug)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Get user by email
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.userStore.GetByEmail(org.ID, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	if err := utils.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token:        token,
		User:         user.ToResponse(),
		Organization: *org,
	}, nil
}

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:         user.ID,
		OrganizationID: user.OrganizationID,
		Role:           user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
