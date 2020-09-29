package universal

import (
	"context"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/xds/auth"
)

func NewNoopAuthenticator() auth.Authenticator {
	return &noopAuthenticator{}
}

type noopAuthenticator struct {
}

func (u *noopAuthenticator) Authenticate(ctx context.Context, dataplane *core_mesh.DataplaneResource, _ auth.Credential) error {
	return nil
}
