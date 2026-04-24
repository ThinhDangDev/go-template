package usecase

import (
	"context"
	"fmt"

	"__MODULE_PATH__/internal/domain"
	"__MODULE_PATH__/internal/domain/entity"
	"__MODULE_PATH__/internal/pkg/authctx"
)

type systemUseCase struct{}

func NewSystemUseCase() *systemUseCase {
	return &systemUseCase{}
}

func (u *systemUseCase) PublicPing(context.Context) string {
	return "public pong"
}

func (u *systemUseCase) RolePing(ctx context.Context, requiredRole string) (*entity.RolePing, error) {
	user, ok := authctx.User(ctx)
	if !ok {
		return nil, domain.ErrMissingCurrentUser
	}

	return &entity.RolePing{
		Message: fmt.Sprintf("%s route is accessible", requiredRole),
		User:    user,
	}, nil
}
