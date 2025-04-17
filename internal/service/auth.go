package service

import (
	"context"

	pb "users/api/auth/v1"
	"users/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type AuthService struct {
	pb.UnimplementedAuthServer
	uc  *biz.AuthUsecase
	log *log.Helper
	t   trace.Tracer
}

func NewAuthService(logger log.Logger, uc *biz.AuthUsecase) *AuthService {
	return &AuthService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *AuthService) Signup(ctx context.Context, req *pb.SignupRequest) (*pb.SignupResponse, error) {
	ctx, span := otel.Tracer("auth").Start(ctx, "AuthService.Signup")
	defer span.End()
	res, err := s.uc.Signup(ctx, req.GetUsername(), req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	resp := &pb.SignupResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}
	return resp, nil
}

func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ctx, span := otel.Tracer("auth").Start(ctx, "AuthService.Login")
	defer span.End()
	res, err := s.uc.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	resp := &pb.LoginResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}
	return resp, nil
}

func (s *AuthService) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	ctx, span := otel.Tracer("auth").Start(ctx, "AuthService.Refresh")
	defer span.End()
	resp := &pb.RefreshResponse{}
	return resp, nil
}

func (s *AuthService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	ctx, span := otel.Tracer("auth").Start(ctx, "AuthService.Logout")
	defer span.End()
	resp := &pb.LogoutResponse{}
	return resp, nil
}

func (s *AuthService) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	ctx, span := otel.Tracer("auth").Start(ctx, "AuthService.Validate")
	defer span.End()
	resp := &pb.ValidateResponse{}
	return resp, nil
}
