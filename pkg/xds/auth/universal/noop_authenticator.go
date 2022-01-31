package universal

import (
	"context"

	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/xds/auth"
)

func NewNoopAuthenticator() auth.Authenticator {
	return &noopAuthenticator{}
}

type noopAuthenticator struct {
}

var _ auth.Authenticator = &noopAuthenticator{}

func (u *noopAuthenticator) Authenticate(ctx context.Context, resource model.Resource, _ auth.Credential) error {
	switch resource := resource.(type) {
	case *core_mesh.DataplaneResource:
		return nil
	case *core_mesh.ZoneIngressResource:
		return nil
	case *core_mesh.ZoneEgressResource:
		return nil
	default:
		return errors.Errorf("no matching authenticator for %s resource", resource.Descriptor().Name)
	}
}
