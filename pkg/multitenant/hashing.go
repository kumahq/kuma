package multitenant

import (
	"context"
)

type Hashing interface {
	ResourceHashKey(ctx context.Context) string
}

type HashingFn func(ctx context.Context) string

func (h HashingFn) ResourceHashKey(ctx context.Context) string {
	return h(ctx)
}

var NoopHashingFn = HashingFn(func(ctx context.Context) string {
	return ""
})
