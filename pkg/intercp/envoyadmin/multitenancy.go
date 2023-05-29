package envoyadmin

import (
	"context"

	"google.golang.org/grpc/metadata"

	"github.com/kumahq/kuma/pkg/multitenant"
)

const TenantMetadataKey = "tenant_id"

func appendTenantMetadata(ctx context.Context) context.Context {
	tenantId, ok := multitenant.TenantFromCtx(ctx)
	if !ok {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, TenantMetadataKey, tenantId)
}

func extractTenantMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	tenantIds := md.Get(TenantMetadataKey)
	if len(tenantIds) != 1 {
		return ctx
	}
	return multitenant.WithTenant(ctx, tenantIds[0])
}
