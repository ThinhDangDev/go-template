package repository

import (
	"context"

	"__MODULE_PATH__/internal/domain/entity"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	Create(ctx context.Context, email, passwordHash, role string) (*entity.User, error)
	List(ctx context.Context) ([]*entity.User, error)
	UpdateAccess(ctx context.Context, id, role string, isActive bool) (*entity.User, error)
	UpsertAdmin(ctx context.Context, email, passwordHash, role string) (*entity.User, error)
}
