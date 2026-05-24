package authenticator

import (
	"context"
	"errors"
	"orchestrator/internal/db"
)

type TokenAuthenticator struct {
	tokenRepository db.TokenRepository
}

func NewTokenAuthenticator(t db.TokenRepository) *TokenAuthenticator {
	return &TokenAuthenticator{
		tokenRepository: t,
	}
}

var _ Authenticator = (*TokenAuthenticator)(nil)

func (a *TokenAuthenticator) Authenticate(ctx context.Context, token string) (string, error) {
	res, err := a.tokenRepository.FindToken(ctx, token)
	if err != nil {
		return "", err
	}

	if !res.Valid {
		return "", errors.New("token has been revoked or is invalid")
	}

	return res.Username, nil
}
