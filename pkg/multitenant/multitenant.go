package multitenant

import (
	"context"
)

type (
	TenantFn  func(ctx context.Context) ([]string, error)
	tenantCtx struct{}
)

func (t TenantFn) GetTenantIds(ctx context.Context) ([]string, error) {
	return t(ctx)
}

func WithTenant(ctx context.Context, tenantId string) context.Context {
	return context.WithValue(ctx, tenantCtx{}, tenantId)
}

func TenantFromCtx(ctx context.Context) string {
	if value, ok := ctx.Value(tenantCtx{}).(string); ok {
		return value
	}
	return ""
}

var SingleTenant = TenantFn(func(ctx context.Context) ([]string, error) {
	return []string{""}, nil
})
