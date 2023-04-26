package multitenant

import (
	"context"
)

type Hashing interface {
	ResourceHashKey(ctx context.Context) string
}

type DefaultHashing struct{}

func (d DefaultHashing) ResourceHashKey(_ context.Context) string {
	return ""
}

var _ Hashing = &DefaultHashing{}
