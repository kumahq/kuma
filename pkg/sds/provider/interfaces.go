package provider

import (
	"context"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
)

type Secret interface {
	ToResource(name string) *envoy_auth.Secret
}

type Identity struct {
	Mesh     string
	Name     string
	Services []string
}

type SecretProvider interface {
	RequiresIdentity() bool
	Get(ctx context.Context, name string, requestor Identity) (Secret, error)
}
