package usecase

import (
	"context"

	"__MODULE_PATH__/internal/domain/entity"
)

type AuthUseCase interface {
	Login(ctx context.Context, email, password string) (*entity.AuthSession, error)
	Me(ctx context.Context) (*entity.User, error)
}
