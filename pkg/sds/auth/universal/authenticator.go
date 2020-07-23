package universal

import (
	"context"

	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	sds_auth "github.com/kumahq/kuma/pkg/sds/auth"
	common_auth "github.com/kumahq/kuma/pkg/sds/auth/common"
	builtin_issuer "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func NewAuthenticator(issuer builtin_issuer.DataplaneTokenIssuer, dataplaneResolver common_auth.DataplaneResolver) sds_auth.Authenticator {
	return &universalAuthenticator{
		issuer:            issuer,
		dataplaneResolver: dataplaneResolver,
	}
}

type universalAuthenticator struct {
	issuer            builtin_issuer.DataplaneTokenIssuer
	dataplaneResolver common_auth.DataplaneResolver
}

func (u *universalAuthenticator) Authenticate(ctx context.Context, proxyId core_xds.ProxyId, credential sds_auth.Credential) (sds_auth.Identity, error) {
	if err := u.reviewToken(proxyId, credential); err != nil {
		return sds_auth.Identity{}, err
	}

	dataplane, err := u.dataplaneResolver(ctx, proxyId)
	if err != nil {
		return sds_auth.Identity{}, errors.Wrapf(err, "unable to find Dataplane for proxy %q", proxyId)
	}
	return common_auth.GetDataplaneIdentity(dataplane)
}

func (u *universalAuthenticator) reviewToken(expectedId core_xds.ProxyId, credential sds_auth.Credential) error {
	proxyId, err := u.issuer.Validate(credential)
	if err != nil {
		return err
	}

	if expectedId.Name != proxyId.Name {
		return errors.Errorf("proxy name from requestor: %s is different than in token: %s", expectedId.Name, proxyId.Name)
	}
	if expectedId.Mesh != proxyId.Mesh {
		return errors.Errorf("proxy mesh from requestor: %s is different than in token: %s", expectedId.Mesh, proxyId.Mesh)
	}
	return nil
}
