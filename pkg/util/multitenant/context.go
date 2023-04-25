package multitenant

import (
	"context"
)

// use custom type like for "userCtx struct{}"
// extract interface
// let's return an array of ContextTenantKey

type Tenant interface {
	GetTenantIds(ctx context.Context) ([]string, error)
	TenantContextKey() any
}

var DefaultTenant = func(ctx context.Context) ([]context.Context, error) {
	return []context.Context{ctx}, nil
}
