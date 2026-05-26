package engine

import (
	"context"
)

type Engine interface {
	Create(ctx context.Context, name string, size string, runtime string, timeout int32, binaryURI string) (string, error)
	Update(ctx context.Context, name string, binaryURI string) error
	Delete(ctx context.Context, name string) error
	List(ctx context.Context) ([]Function, error)
}

type Function struct {
	Name       string
	Runtime    string
	MemorySize int32
	Url        string
	Status     string
}
