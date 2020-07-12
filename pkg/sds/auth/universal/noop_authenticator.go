package universal

import (
	"context"

	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	sds_auth "github.com/kumahq/kuma/pkg/sds/auth"
	common_auth "github.com/kumahq/kuma/pkg/sds/auth/common"
)

func NewNoopAuthenticator(dataplaneResolver common_auth.DataplaneResolver) sds_auth.Authenticator {
	return &noopAuthenticator{
		dataplaneResolver: dataplaneResolver,
	}
}

type noopAuthenticator struct {
	dataplaneResolver common_auth.DataplaneResolver
}

func (u *noopAuthenticator) Authenticate(ctx context.Context, proxyId core_xds.ProxyId, _ sds_auth.Credential) (sds_auth.Identity, error) {
	dataplane, err := u.dataplaneResolver(ctx, proxyId)
	if err != nil {
		return sds_auth.Identity{}, errors.Wrapf(err, "unable to find Dataplane for proxy %q", proxyId)
	}
	return common_auth.GetDataplaneIdentity(dataplane)
}
