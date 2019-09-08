package universal

import (
	"context"

	"github.com/pkg/errors"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	common_auth "github.com/Kong/kuma/pkg/sds/auth/common"
)

func New(dataplaneResolver common_auth.DataplaneResolver) sds_auth.Authenticator {
	return &universalAuthenticator{
		dataplaneResolver: dataplaneResolver,
	}
}

type universalAuthenticator struct {
	dataplaneResolver common_auth.DataplaneResolver
}

func (u *universalAuthenticator) Authenticate(ctx context.Context, proxyId core_xds.ProxyId, credential sds_auth.Credential) (sds_auth.Identity, error) {
	dataplane, err := u.dataplaneResolver(ctx, proxyId)
	if err != nil {
		return sds_auth.Identity{}, errors.Wrapf(err, "unable to find Dataplane for proxy %q", proxyId)
	}
	return common_auth.GetDataplaneIdentity(dataplane)
}
