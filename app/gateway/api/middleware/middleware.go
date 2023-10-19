package middleware

import (
	"context"
	"time"
)

//go:generate moq -rm -out middleware_mocks.gen.go . cache

type cache interface {
	Exists(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, obj any, ttl time.Duration) error
}
