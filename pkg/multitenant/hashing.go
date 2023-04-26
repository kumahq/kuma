package multitenant

import (
	"context"
)

type Hashing interface {
	ResourceHashKey(ctx context.Context) string
	KdsHashFn(ctx context.Context, id string) string
	SinkStatusCacheKey(ctx context.Context) string
}

type DefaultHashing struct{}

func (d DefaultHashing) ResourceHashKey(_ context.Context) string {
	return ""
}

var _ Hashing = &DefaultHashing{}

func (d DefaultHashing) KdsHashFn(_ context.Context, id string) string {
	return id
}

func (d DefaultHashing) SinkStatusCacheKey(_ context.Context) string {
	return ""
}
