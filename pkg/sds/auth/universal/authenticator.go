package universal

import (
	"context"
	"github.com/dgrijalva/jwt-go"

	"github.com/pkg/errors"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	common_auth "github.com/Kong/kuma/pkg/sds/auth/common"
)

func NewAuthenticator(privateKey []byte, dataplaneResolver common_auth.DataplaneResolver) sds_auth.Authenticator {
	return &universalAuthenticator{
		privateKey:        privateKey,
		dataplaneResolver: dataplaneResolver,
	}
}

type universalAuthenticator struct {
	privateKey        []byte
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

func (u *universalAuthenticator) reviewToken(proxyId core_xds.ProxyId, credential sds_auth.Credential) error {
	c := &claims{}

	token, err := jwt.ParseWithClaims(string(credential), c, func(*jwt.Token) (interface{}, error) {
		return u.privateKey, nil
	})
	if err != nil {
		return errors.Wrap(err, "could not parse token")
	}
	if !token.Valid {
		return errors.New("token is not valid")
	}

	if proxyId.Name != c.Name {
		return errors.Errorf("proxy name from requestor is different than in token. Expected %s got %s", proxyId.Name, c.Name)
	}
	if proxyId.Mesh != c.Mesh {
		return errors.Errorf("proxy mesh from requestor is different than in token. Expected %s got %s", proxyId.Mesh, c.Mesh)
	}
	return nil
}
