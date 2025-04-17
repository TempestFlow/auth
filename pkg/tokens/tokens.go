package tokens

import (
	"errors"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type TokenFactory struct {
	name        string
	secret      string
	access_exp  time.Duration
	refresh_exp time.Duration
}

func NewTokenFactory(name, secret string) *TokenFactory {
	return &TokenFactory{
		name:   name,
		secret: secret,
	}
}

type TokenPayload struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	ExtraClaims map[string]interface{} `json:"extra_claims"`

	builder *jwt.Builder
	tf      *TokenFactory
}

func (p TokenPayload) Build(duration time.Duration) TokenPayload {
	builder := jwt.NewBuilder().IssuedAt(time.Now()).
		NotBefore(time.Now()).
		Issuer(p.tf.name).
		Expiration(time.Now().Add(duration)).
		Subject(p.ID).
		Claim("username", p.Username).
		Claim("email", p.Email).
		Claim("extraClaims", p.ExtraClaims)

	p.builder = builder
	return p
}

func (p TokenPayload) Sign() ([]byte, error) {
	if p.builder == nil {
		return nil, errors.New("builder is nil")
	}
	if p.tf.secret == "" {
		return nil, errors.New("secret is empty")
	}
	token, err := p.builder.Build()
	if err != nil {
		return nil, err
	}
	opts := []jwt.SignOption{jwt.WithKey(jwa.HS256(), p.tf.secret)}
	tok, err := jwt.Sign(token, opts...)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func (p TokenPayload) Parse(token string) (*TokenPayload, error) {
	tok := []byte(token)
	payload, err := jwt.Parse(tok)
	if err != nil {
		return nil, err
	}
	id, ok := payload.Subject()
	if !ok {
		return nil, errors.New("subject is not string")
	}
	var username string
	err = payload.Get("username", &username)
	if err != nil {
		return nil, err
	}

	var email string
	err = payload.Get("email", &email)
	if err != nil {
		return nil, err
	}

	var extraClaims map[string]interface{}
	err = payload.Get("extraClaims", &extraClaims)
	if err != nil {
		return nil, err
	}

	return &TokenPayload{
		ID:          id,
		Username:    username,
		Email:       email,
		ExtraClaims: extraClaims,
		builder:     nil,
		tf:          p.tf,
	}, nil
}

func Validate(token jwt.Token) error {
	return jwt.Validate(token)
}
