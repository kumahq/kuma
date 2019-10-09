package auth

import (
	"context"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

type Credential string

type Identity struct {
	Mesh    string
	Service string
}

type Authenticator interface {
	Authenticate(ctx context.Context, proxyId core_xds.ProxyId, credential Credential) (Identity, error)
}

type CredentialGenerator interface {
	Generate(proxyId core_xds.ProxyId) (Credential, error)
}
