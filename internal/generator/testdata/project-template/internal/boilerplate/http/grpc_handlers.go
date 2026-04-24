package transport

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"__MODULE_PATH__/internal/boilerplate/app"
	"__MODULE_PATH__/internal/boilerplate/auth"
	"__MODULE_PATH__/internal/boilerplate/store"
	pb "__MODULE_PATH__/protogen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TemplateService struct {
	pb.UnimplementedTemplateServiceServer
	runtime *app.Runtime
}

func NewTemplateService(runtime *app.Runtime) *TemplateService {
	return &TemplateService{runtime: runtime}
}

func (s *TemplateService) PublicPing(context.Context, *emptypb.Empty) (*pb.PingResponse, error) {
	return &pb.PingResponse{Message: "public pong"}, nil
}

func (s *TemplateService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	email := strings.TrimSpace(req.GetEmail())
	password := strings.TrimSpace(req.GetPassword())
	if email == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.runtime.Users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		s.runtime.Logger.Error("failed to load user", "error", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}
	if !user.IsActive {
		return nil, status.Error(codes.PermissionDenied, "user is inactive")
	}

	if err := auth.ComparePassword(user.PasswordHash, password); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, err := s.runtime.Tokens.IssueAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		s.runtime.Logger.Error("failed to issue jwt", "error", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.runtime.Config.JWTAccessTTL.Seconds()),
		User:        protoUser(user),
	}, nil
}

func (s *TemplateService) Me(ctx context.Context, _ *emptypb.Empty) (*pb.MeResponse, error) {
	user, ok := getCurrentUserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing current user")
	}

	return &pb.MeResponse{User: protoUser(user)}, nil
}

func (s *TemplateService) AdminPing(ctx context.Context, _ *emptypb.Empty) (*pb.RolePingResponse, error) {
	return rolePingResponse(ctx, "admin")
}

func (s *TemplateService) OperatorPing(ctx context.Context, _ *emptypb.Empty) (*pb.RolePingResponse, error) {
	return rolePingResponse(ctx, "operator")
}

func (s *TemplateService) ViewerPing(ctx context.Context, _ *emptypb.Empty) (*pb.RolePingResponse, error) {
	return rolePingResponse(ctx, "viewer")
}

func rolePingResponse(ctx context.Context, requiredRole string) (*pb.RolePingResponse, error) {
	user, ok := getCurrentUserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing current user")
	}

	return &pb.RolePingResponse{
		Message: fmt.Sprintf("%s route is accessible", requiredRole),
		User:    protoUser(user),
	}, nil
}

func protoUser(user *store.User) *pb.User {
	if user == nil {
		return nil
	}

	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
