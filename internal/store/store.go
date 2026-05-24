package store

import (
	"context"
)

type Store interface {
	GenerateUploadURL(ctx context.Context, key string) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, key string) error
}
