package multitenant

import (
	"context"
	"errors"
)

// GlobalTenantID is a unique ID used for storing resources that are not tenant-aware
var GlobalTenantID = ""

type tenantCtx struct{}

type Tenants interface {
	// GetID gets id of tenant from context.
	// Design: why not rely on TenantFromCtx? Different implementations of Tenants can have different error handling.
	// Some may return error on missing tenant, whereas Kuma never requires tenant set in context.
	GetID(ctx context.Context) (string, error)
	GetIDs(ctx context.Context) ([]string, error)
}

var SingleTenant = &singleTenant{}

type singleTenant struct{}

func (s singleTenant) GetID(context.Context) (string, error) {
	return "", nil
}

func (s singleTenant) GetIDs(context.Context) ([]string, error) {
	return []string{""}, nil
}

func WithTenant(ctx context.Context, tenantId string) context.Context {
	return context.WithValue(ctx, tenantCtx{}, tenantId)
}

func TenantFromCtx(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(tenantCtx{}).(string)
	return value, ok
}

// CopyIntoCtx copies tenant information from src context to dst context
func CopyIntoCtx(src context.Context, dst context.Context) context.Context {
	tenantId, ok := TenantFromCtx(src)
	if !ok {
		return dst
	}
	return WithTenant(dst, tenantId)
}

var TenantMissingErr = errors.New("tenant is missing")
