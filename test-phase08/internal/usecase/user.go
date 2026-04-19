package usecase

import (
	"context"
	"errors"

	"github.com/user/test-phase08/internal/domain/entity"
	"github.com/user/test-phase08/internal/domain/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)

// UserUsecase handles user business logic
type UserUsecase struct {
	repo repository.UserRepository
}

// NewUserUsecase creates a new user usecase
func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// Create creates a new user with hashed password
func (u *UserUsecase) Create(ctx context.Context, user *entity.User) error {
	// Check if email exists
	existing, _ := u.repo.GetByEmail(ctx, user.Email)
	if existing != nil {
		return ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return u.repo.Create(ctx, user)
}

// GetByID retrieves a user by ID
func (u *UserUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return u.repo.GetByID(ctx, id)
}

// GetByEmail retrieves a user by email
func (u *UserUsecase) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return u.repo.GetByEmail(ctx, email)
}

// List retrieves users with pagination
func (u *UserUsecase) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	return u.repo.List(ctx, limit, offset)
}

// Update updates user information
func (u *UserUsecase) Update(ctx context.Context, user *entity.User) error {
	existing, err := u.repo.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrUserNotFound
	}

	return u.repo.Update(ctx, user)
}

// Delete soft deletes a user
func (u *UserUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

// VerifyPassword verifies user password
func (u *UserUsecase) VerifyPassword(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}
