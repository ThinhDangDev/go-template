package usecase

import (
	"context"

	"__MODULE_PATH__/internal/domain/entity"
)

type SystemUseCase interface {
	PublicPing(ctx context.Context) string
	RolePing(ctx context.Context, requiredRole string) (*entity.RolePing, error)
}
