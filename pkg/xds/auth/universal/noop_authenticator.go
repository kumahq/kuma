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

func (u *noopAuthenticator) AuthenticateZoneIngress(ctx context.Context, zoneIngress *core_mesh.ZoneIngressResource, _ auth.Credential) error {
	return nil
}
