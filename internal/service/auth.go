package service

import (
	"context"
	"time"

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

	res, err := s.uc.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}
	resp := &pb.RefreshResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}
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

	res, err := s.uc.Validate(ctx, req.GetAccessToken())
	if err != nil {
		return nil, err
	}
	expHuman := time.Unix(res.Exp, 0).Format(time.RFC3339)
	resp := &pb.ValidateResponse{
		Username: res.Username,
		Email:    res.Email,
		Id:       res.ID,
		Valid:    res.Valid,
		Exp:      expHuman,
	}

	return resp, nil
}
