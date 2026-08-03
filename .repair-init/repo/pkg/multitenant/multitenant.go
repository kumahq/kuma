package multitenant

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const spanAttributeName = "tenantID"

// GlobalTenantID is a unique ID used for storing resources that are not tenant-aware
var GlobalTenantID = ""

type tenantCtx struct{}

type Tenants interface {
	// GetID gets id of tenant from context.
	// Design: why not rely on TenantFromCtx? Different implementations of Tenants can have different error handling.
	// Some may return error on missing tenant, whereas Kuma never requires tenant set in context.
	GetID(ctx context.Context) (string, error)
	GetIDs(ctx context.Context) ([]string, error)

	// SupportsSharding returns true if Tenants implementation supports sharding.
	// It means that GetIDs will return only subset of tenants.
	SupportsSharding() bool
	// IDSupported returns true if given tenant is in the current shard.
	IDSupported(ctx context.Context, id string) (bool, error)
}

var SingleTenant = &singleTenant{}

type singleTenant struct{}

func (s singleTenant) GetID(context.Context) (string, error) {
	return "", nil
}

func (s singleTenant) GetIDs(context.Context) ([]string, error) {
	return []string{""}, nil
}

func (s singleTenant) SupportsSharding() bool {
	return false
}

func (s singleTenant) IDSupported(ctx context.Context, id string) (bool, error) {
	return true, nil
}

func WithTenant(ctx context.Context, tenantId string) context.Context {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetAttributes(attribute.String(spanAttributeName, tenantId))
	}

	return context.WithValue(ctx, tenantCtx{}, tenantId)
}

func TenantFromCtx(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(tenantCtx{}).(string)
	return value, ok
}

var ErrTenantMissing = errors.New("tenant is missing")
