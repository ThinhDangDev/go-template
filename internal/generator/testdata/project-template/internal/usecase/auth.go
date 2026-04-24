package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"__MODULE_PATH__/internal/boilerplate/auth"
	"__MODULE_PATH__/internal/domain"
	"__MODULE_PATH__/internal/domain/entity"
	repositorycontract "__MODULE_PATH__/internal/domain/interface/repository"
	"__MODULE_PATH__/internal/pkg/authctx"
)

type authUseCase struct {
	users     repositorycontract.UserRepository
	tokens    *auth.TokenManager
	logger    *slog.Logger
	accessTTL time.Duration
}

func NewAuthUseCase(
	users repositorycontract.UserRepository,
	tokens *auth.TokenManager,
	logger *slog.Logger,
	accessTTL time.Duration,
) *authUseCase {
	return &authUseCase{
		users:     users,
		tokens:    tokens,
		logger:    logger,
		accessTTL: accessTTL,
	}
}

func (u *authUseCase) Register(ctx context.Context, email, password string) (*entity.AuthSession, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return nil, domain.ErrInvalidInput
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		u.logger.Error("failed to hash password", "error", err)
		return nil, err
	}

	user, err := u.users.Create(ctx, email, passwordHash, entity.RoleViewer)
	if err != nil {
		return nil, err
	}

	return u.issueSession(user)
}

func (u *authUseCase) Login(ctx context.Context, email, password string) (*entity.AuthSession, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, domain.ErrInactiveUser
	}

	if err := auth.ComparePassword(user.PasswordHash, password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return u.issueSession(user)
}

func (u *authUseCase) issueSession(user *entity.User) (*entity.AuthSession, error) {
	token, err := u.tokens.IssueAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		u.logger.Error("failed to issue jwt", "error", err)
		return nil, err
	}

	return &entity.AuthSession{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(u.accessTTL.Seconds()),
		User:        user,
	}, nil
}

func (u *authUseCase) Me(ctx context.Context) (*entity.User, error) {
	user, ok := authctx.User(ctx)
	if !ok {
		return nil, domain.ErrMissingCurrentUser
	}

	return user, nil
}
