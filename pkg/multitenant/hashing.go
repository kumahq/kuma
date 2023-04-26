package multitenant

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type Hashing interface {
	GetOptionsFunc() store.GetOptionsFunc
	ListOptionsFunc() store.ListOptionsFunc
	KdsHashFn(ctx context.Context, id string) string
	SinkStatusCacheKey(ctx context.Context) string
}

type DefaultHashing struct{}

var _ Hashing = &DefaultHashing{}

func (d DefaultHashing) GetOptionsFunc() store.GetOptionsFunc {
	return func(options *store.GetOptions) {
		options.KeyFromContext = func(ctx context.Context) string {
			return "global"
		}
	}
}

func (d DefaultHashing) ListOptionsFunc() store.ListOptionsFunc {
	return func(options *store.ListOptions) {
		options.Suffix = func(ctx context.Context) string {
			return "global"
		}
	}
}

func (d DefaultHashing) KdsHashFn(_ context.Context, id string) string {
	return id
}

func (d DefaultHashing) SinkStatusCacheKey(_ context.Context) string {
	return ""
}
