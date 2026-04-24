package transport

import (
	"context"
	"fmt"
	"strings"

	"__MODULE_PATH__/internal/pkg/authctx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcRouteBinding struct {
	Public bool
	Path   string
	Method string
}

var grpcRouteBindings = map[string]grpcRouteBinding{
	"/boilerplate.v1.TemplateService/PublicPing": {Public: true},
	"/boilerplate.v1.TemplateService/Register":   {Public: true},
	"/boilerplate.v1.TemplateService/Login":      {Public: true},
	"/boilerplate.v1.TemplateService/Me":         {Path: "/api/v1/auth/me", Method: "GET"},
	"/boilerplate.v1.TemplateService/AdminPing":  {Path: "/api/v1/admin/ping", Method: "GET"},
	"/boilerplate.v1.TemplateService/ListUsers":  {Path: "/api/v1/admin/users", Method: "GET"},
	"/boilerplate.v1.TemplateService/ListRoles":  {Path: "/api/v1/admin/roles", Method: "GET"},
	"/boilerplate.v1.TemplateService/UpdateUserAccess": {
		Path:   "/api/v1/admin/users/:user_id/access",
		Method: "PATCH",
	},
}

func (s *Server) AuthUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		binding, ok := grpcRouteBindings[info.FullMethod]
		if !ok || binding.Public {
			return handler(ctx, req)
		}

		token, err := grpcBearerToken(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		claims, err := s.runtime.Tokens.ValidateToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		user, err := s.runtime.Users.GetByID(ctx, claims.UserID)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "user not found")
		}
		if !user.IsActive {
			return nil, status.Error(codes.PermissionDenied, "user is inactive")
		}

		claims.Email = user.Email
		claims.Role = user.Role

		allowed, err := s.runtime.Authorizer.Authorize(claims.Role, binding.Path, binding.Method)
		if err != nil {
			s.runtime.Logger.Error("grpc rbac authorization failed", "error", err, "method", info.FullMethod)
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}
		if !allowed {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		ctx = authctx.WithAuth(ctx, claims, user)
		return handler(ctx, req)
	}
}

func grpcBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", fmt.Errorf("missing bearer token")
	}

	token, err := extractBearerToken(values[0])
	if err != nil {
		return "", err
	}

	return token, nil
}

func extractBearerToken(header string) (string, error) {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", fmt.Errorf("missing bearer token")
	}

	return parts[1], nil
}
