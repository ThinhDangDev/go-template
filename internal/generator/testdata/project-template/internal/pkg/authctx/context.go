package authctx

import (
	"context"

	"__MODULE_PATH__/internal/boilerplate/auth"
	"__MODULE_PATH__/internal/domain/entity"
)

type contextKey string

const (
	claimsKey contextKey = "request.auth_claims"
	userKey   contextKey = "request.auth_user"
)

func WithAuth(ctx context.Context, claims *auth.Claims, user *entity.User) context.Context {
	ctx = context.WithValue(ctx, claimsKey, claims)
	ctx = context.WithValue(ctx, userKey, user)
	return ctx
}

func Claims(ctx context.Context) (*auth.Claims, bool) {
	value := ctx.Value(claimsKey)
	claims, ok := value.(*auth.Claims)
	return claims, ok
}

func User(ctx context.Context) (*entity.User, bool) {
	value := ctx.Value(userKey)
	user, ok := value.(*entity.User)
	return user, ok
}
