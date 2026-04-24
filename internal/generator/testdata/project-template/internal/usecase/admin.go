package usecase

import (
	"context"
	"strings"

	"__MODULE_PATH__/internal/domain"
	"__MODULE_PATH__/internal/domain/entity"
	repositorycontract "__MODULE_PATH__/internal/domain/interface/repository"
)

type adminUseCase struct {
	users repositorycontract.UserRepository
}

func NewAdminUseCase(users repositorycontract.UserRepository) *adminUseCase {
	return &adminUseCase{
		users: users,
	}
}

func (u *adminUseCase) ListUsers(ctx context.Context) ([]*entity.User, error) {
	return u.users.List(ctx)
}

func (u *adminUseCase) ListRoles(_ context.Context) []entity.RoleDefinition {
	return entity.AvailableRoles()
}

func (u *adminUseCase) UpdateUserAccess(ctx context.Context, userID, role string, isActive bool) (*entity.User, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domain.ErrInvalidInput
	}

	role = entity.NormalizeRole(role)
	if !entity.IsValidRole(role) {
		return nil, domain.ErrInvalidRole
	}

	return u.users.UpdateAccess(ctx, userID, role, isActive)
}
