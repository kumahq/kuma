package auth

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type perProxyTypeAuthenticator struct {
	dpProxyAuthenticator   Authenticator
	zoneProxyAuthenticator Authenticator
}

var _ Authenticator = &perProxyTypeAuthenticator{}

func NewPerProxyTypeAuthenticator(dpProxyAuthenticator, zoneProxyAuthenticator Authenticator) Authenticator {
	return &perProxyTypeAuthenticator{
		dpProxyAuthenticator:   dpProxyAuthenticator,
		zoneProxyAuthenticator: zoneProxyAuthenticator,
	}
}

func (p perProxyTypeAuthenticator) Authenticate(ctx context.Context, resource core_model.Resource, credential Credential) error {
	switch resource.Descriptor().Name {
	case mesh.ZoneIngressType, mesh.ZoneEgressType:
		return p.zoneProxyAuthenticator.Authenticate(ctx, resource, credential)
	case mesh.DataplaneType:
		return p.dpProxyAuthenticator.Authenticate(ctx, resource, credential)
	default:
		return errors.New("unknown proxy type")
	}
}
