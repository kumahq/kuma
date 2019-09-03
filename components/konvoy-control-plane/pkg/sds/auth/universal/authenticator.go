package universal

import (
	"context"

	"github.com/pkg/errors"

	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	sds_auth "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/sds/auth"
)

func New() sds_auth.Authenticator {
	return universalAuthenticator{}
}

type universalAuthenticator struct {
}

func (_ universalAuthenticator) Authenticate(ctx context.Context, proxyId core_xds.ProxyId, credential sds_auth.Credential) (sds_auth.Identity, error) {
	return sds_auth.Identity{}, errors.New("not implemented")
}
