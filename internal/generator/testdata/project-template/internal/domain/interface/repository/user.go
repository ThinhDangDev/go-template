package repository

import (
	"context"

	"__MODULE_PATH__/internal/domain/entity"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	UpsertAdmin(ctx context.Context, email, passwordHash, role string) (*entity.User, error)
}
