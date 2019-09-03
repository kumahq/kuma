package provider

import (
	"context"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"

	sds_auth "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/sds/auth"
)

type Secret interface {
	ToResource(name string) *envoy_auth.Secret
}

type SecretProvider interface {
	RequiresIdentity() bool
	Get(ctx context.Context, name string, requestor sds_auth.Identity) (Secret, error)
}
