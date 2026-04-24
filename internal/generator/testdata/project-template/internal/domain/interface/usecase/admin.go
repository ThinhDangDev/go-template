package usecase

import (
	"context"

	"__MODULE_PATH__/internal/domain/entity"
)

type AdminUseCase interface {
	ListUsers(ctx context.Context) ([]*entity.User, error)
	ListRoles(ctx context.Context) []entity.RoleDefinition
	UpdateUserAccess(ctx context.Context, userID, role string, isActive bool) (*entity.User, error)
}
