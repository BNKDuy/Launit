package db

import (
	"context"
	"orchestrator/internal/domain"
)

type TokenRepository interface {
	SaveToken(ctx context.Context, token domain.Token) error
	FindToken(ctx context.Context, value string) (domain.Token, error)
}
