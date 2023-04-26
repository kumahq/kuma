package multitenant

import (
	"context"
)

type HashingFn func(ctx context.Context) string

func (h HashingFn) ResourceHashKey(ctx context.Context) string {
	return h(ctx)
}

var DefaultHashingFn = HashingFn(func(ctx context.Context) string {
	return ""
})

