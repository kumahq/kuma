package universal

import (
	"context"

	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/xds/auth"
)

func NewNoopAuthenticator() auth.Authenticator {
	return &noopAuthenticator{}
}

type noopAuthenticator struct{}

var _ auth.Authenticator = &noopAuthenticator{}

func (u *noopAuthenticator) Authenticate(ctx context.Context, resource model.Resource, _ auth.Credential) (auth.Binding, error) {
	switch resource := resource.(type) {
	case *core_mesh.DataplaneResource, *core_mesh.ZoneIngressResource, *core_mesh.ZoneEgressResource:
		return auth.Binding{Unauthenticated: true}, nil
	default:
		return auth.Binding{}, errors.Errorf("no matching authenticator for %s resource", resource.Descriptor().Name)
	}
}
