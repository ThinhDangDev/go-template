package handler

import (
	"context"
	"errors"

	"__MODULE_PATH__/internal/domain"
	"__MODULE_PATH__/internal/domain/entity"
	usecasecontract "__MODULE_PATH__/internal/domain/interface/usecase"
	pb "__MODULE_PATH__/protogen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TemplateService struct {
	pb.UnimplementedTemplateServiceServer
	authUseCase   usecasecontract.AuthUseCase
	systemUseCase usecasecontract.SystemUseCase
}

func NewTemplateService(
	authUseCase usecasecontract.AuthUseCase,
	systemUseCase usecasecontract.SystemUseCase,
) *TemplateService {
	return &TemplateService{
		authUseCase:   authUseCase,
		systemUseCase: systemUseCase,
	}
}

func (h *TemplateService) PublicPing(ctx context.Context, _ *emptypb.Empty) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		Message: h.systemUseCase.PublicPing(ctx),
	}, nil
}

func (h *TemplateService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	session, err := h.authUseCase.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapAuthError(err)
	}

	return &pb.LoginResponse{
		AccessToken: session.AccessToken,
		TokenType:   session.TokenType,
		ExpiresIn:   session.ExpiresIn,
		User:        protoUser(session.User),
	}, nil
}

func (h *TemplateService) Me(ctx context.Context, _ *emptypb.Empty) (*pb.MeResponse, error) {
	user, err := h.authUseCase.Me(ctx)
	if err != nil {
		return nil, mapAuthError(err)
	}

	return &pb.MeResponse{User: protoUser(user)}, nil
}

func (h *TemplateService) AdminPing(ctx context.Context, _ *emptypb.Empty) (*pb.RolePingResponse, error) {
	return h.rolePing(ctx, "admin")
}

func (h *TemplateService) OperatorPing(ctx context.Context, _ *emptypb.Empty) (*pb.RolePingResponse, error) {
	return h.rolePing(ctx, "operator")
}

func (h *TemplateService) ViewerPing(ctx context.Context, _ *emptypb.Empty) (*pb.RolePingResponse, error) {
	return h.rolePing(ctx, "viewer")
}

func (h *TemplateService) rolePing(ctx context.Context, requiredRole string) (*pb.RolePingResponse, error) {
	result, err := h.systemUseCase.RolePing(ctx, requiredRole)
	if err != nil {
		return nil, mapAuthError(err)
	}

	return &pb.RolePingResponse{
		Message: result.Message,
		User:    protoUser(result.User),
	}, nil
}

func protoUser(user *entity.User) *pb.User {
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

func mapAuthError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, domain.ErrInactiveUser):
		return status.Error(codes.PermissionDenied, "user is inactive")
	case errors.Is(err, domain.ErrMissingCurrentUser):
		return status.Error(codes.Unauthenticated, "missing current user")
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
