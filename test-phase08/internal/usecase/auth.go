
package usecase

import (
	"context"
	"errors"

	"github.com/user/test-phase08/internal/domain/entity"
	"github.com/user/test-phase08/internal/domain/repository"
	"github.com/user/test-phase08/internal/infrastructure/auth"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

// AuthUsecase handles authentication business logic
type AuthUsecase struct {
	userRepo repository.UserRepository
	jwtService *auth.JWTService
	userUsecase *UserUsecase
}

// NewAuthUsecase creates a new auth usecase
func NewAuthUsecase(
	userRepo repository.UserRepository,
	jwtService *auth.JWTService,
	userUsecase *UserUsecase,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo,
		jwtService: jwtService,
		userUsecase: userUsecase,
	}
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}
// LoginResponse represents login response with tokens
type LoginResponse struct {
	User   *entity.User      `json:"user"`
	Tokens *auth.TokenPair   `json:"tokens"`
}

// Login authenticates user and returns tokens
func (u *AuthUsecase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := u.userUsecase.VerifyPassword(ctx, req.Email, req.Password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.Active {
		return nil, errors.New("user account is inactive")
	}

	tokens, err := u.jwtService.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		User:   user,
		Tokens: tokens,
	}, nil
}

// Register creates new user and returns tokens
func (u *AuthUsecase) Register(ctx context.Context, req RegisterRequest) (*LoginResponse, error) {
	// Check if user exists
	existing, _ := u.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrUserExists
	}

	// Create user
	user := &entity.User{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Role:     "user",
		Active:   true,
	}

	if err := u.userUsecase.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	tokens, err := u.jwtService.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		User:   user,
		Tokens: tokens,
	}, nil
}

// RefreshToken creates new access token from refresh token
func (u *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return u.jwtService.RefreshAccessToken(refreshToken)
}
