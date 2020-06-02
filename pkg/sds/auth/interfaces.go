package auth

import (
	"context"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

type Credential string

type Identity struct {
	Mesh     string
	Services []string
}

type Authenticator interface {
	Authenticate(ctx context.Context, proxyId core_xds.ProxyId, credential Credential) (Identity, error)
}
