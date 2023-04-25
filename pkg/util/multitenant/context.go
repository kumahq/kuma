package multitenant

import (
	"context"
)

// use custom type like for "userCtx struct{}"
// extract interface
// let's return an array of ContextTenantKey

var DefaultTenant = func(ctx context.Context) ([]context.Context, error) {
	return []context.Context{ctx}, nil
}
