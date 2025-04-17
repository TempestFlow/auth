package biz

import (
	"context"
	"time"

	"users/internal/conf"
	"users/pkg/tokens"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	t       trace.Tracer
	usersUC *UsersUsecase
	log     *log.Helper
	tf      *tokens.TokenFactory
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewAuthUsecase(usersUC *UsersUsecase, logger log.Logger, conf *conf.Bootstrap) *AuthUsecase {
	tracer := otel.GetTracerProvider().Tracer("AuthUsecase")
	name := conf.Metadata.Name
	// secret := conf.JWT.Secret
	secret := "secret"
	tf := tokens.NewTokenFactory(name, secret)
	return &AuthUsecase{
		usersUC: usersUC,
		log:     log.NewHelper(logger),
		t:       tracer,
		tf:      tf,
	}
}

func (uc *AuthUsecase) Signup(ctx context.Context, username, email, password string) (*TokenPair, error) {
	ctx, span := uc.t.Start(ctx, "Signup")
	defer span.End()

	userID, err := uc.usersUC.CreateUser(ctx, &User{
		Username: username,
		Email:    email,
		Password: password,
		Phone:    "1234567890",
	})
	if err != nil {
		uc.log.Error("signup: failed to create user", err)
		return nil, err
	}

	accToken, err := uc.tf.NewTokenPayload().
		SetID(userID).
		SetEmail(email).
		SetUsername(username).
		Build(time.Minute * 5).
		Sign()
	if err != nil {
		uc.log.Error(err)
		return nil, err
	}
	refToken, err := uc.tf.NewTokenPayload().
		SetID(userID).
		SetEmail(email).
		SetUsername(username).
		Build(time.Hour * 24).
		Sign()
	if err != nil {
		uc.log.Error(err)
		return nil, err
	}

	return &TokenPair{
		AccessToken:  string(accToken),
		RefreshToken: string(refToken),
	}, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	ctx, span := uc.t.Start(ctx, "Login")
	defer span.End()

	user, err := uc.usersUC.GetUserByUsername(ctx, username)
	if err != nil {
		uc.log.Error("login: failed to get user", err)
		return nil, err
	}
	if user == nil {
		uc.log.Error("login: user not found")
		return nil, errors.NotFound("user", "user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		uc.log.Error("login: failed to compare password", err)
		return nil, err
	}

	accToken, err := uc.tf.NewTokenPayload().
		SetID(user.ID).
		SetEmail(user.Email).
		SetUsername(user.Username).
		Build(time.Minute * 5).
		Sign()
	if err != nil {
		uc.log.Error(err)
		return nil, err
	}
	refToken, err := uc.tf.NewTokenPayload().
		SetID(user.ID).
		SetEmail(user.Email).
		SetUsername(user.Username).
		Build(time.Hour * 24).
		Sign()
	if err != nil {
		uc.log.Error(err)
		return nil, err
	}

	return &TokenPair{
		AccessToken:  string(accToken),
		RefreshToken: string(refToken),
	}, nil
}
