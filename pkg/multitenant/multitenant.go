package multitenant

import (
	"context"
)

type Tenant interface {
	GetTenantIds(ctx context.Context) ([]string, error)
	TenantContextKey() any
}

type DefaultTenant struct{}

var _ Tenant = &DefaultTenant{}

func (d DefaultTenant) GetTenantIds(_ context.Context) ([]string, error) {
	return []string{""}, nil
}

type tenantCtx struct{}

func (d DefaultTenant) TenantContextKey() any {
	return tenantCtx{}
}
