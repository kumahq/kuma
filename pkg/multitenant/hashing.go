package multitenant

import (
	"context"
)

type Hashing func(ctx context.Context) string

func (h Hashing) ResourceHashKey(ctx context.Context) string {
	return h(ctx)
}

var DefaultHashingFn = Hashing(func(ctx context.Context) string {
	return ""
})

