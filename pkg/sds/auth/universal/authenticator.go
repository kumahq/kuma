package universal

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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

// universalAuthenticator defines authentication for Dataplane Tokens
// All fields in token are optional, so we only validate data that is available in token. This way you can pick your level of security.
// Generate token for mesh+name for maximum security.
// Generate token for mesh+tags(ex. kuma.io/service) so you can reuse the token for many instances of the same service.
// Generate token for mesh if you trust the scope of the mesh.
//
// If you generate token bound to tags all tags values have to match the dataplane, so for example if you have a Dataplane
// with inbounds: 1) kuma.io/service:web 2) kuma.io/service:web-api, you need token for both values kuma.io/service=web,web-api
// Dataplane also needs to have all tags defined in the token
type universalAuthenticator struct {
	issuer            builtin_issuer.DataplaneTokenIssuer
	dataplaneResolver common_auth.DataplaneResolver
}

func (u *universalAuthenticator) Authenticate(ctx context.Context, proxyId core_xds.ProxyId, credential sds_auth.Credential) (sds_auth.Identity, error) {
	dataplane, err := u.dataplaneResolver(ctx, proxyId)
	if err != nil {
		return sds_auth.Identity{}, errors.Wrapf(err, "unable to find Dataplane for proxy %q", proxyId)
	}
	if err := u.reviewToken(dataplane, credential); err != nil {
		return sds_auth.Identity{}, err
	}
	return common_auth.GetDataplaneIdentity(dataplane)
}

func (u *universalAuthenticator) reviewToken(dataplane *mesh.DataplaneResource, credential sds_auth.Credential) error {
	dpIdentity, err := u.issuer.Validate(credential)
	if err != nil {
		return err
	}

	if dpIdentity.Name != "" {
		if dataplane.Meta.GetName() != dpIdentity.Name {
			return errors.Errorf("proxy name from requestor: %s is different than in token: %s", dataplane.Meta.GetName(), dpIdentity.Name)
		}
	}
	if dpIdentity.Mesh != "" {
		if dataplane.Meta.GetMesh() != dpIdentity.Mesh {
			return errors.Errorf("proxy mesh from requestor: %s is different than in token: %s", dataplane.Meta.GetMesh(), dpIdentity.Mesh)
		}
	}
	if err := validateTags(dpIdentity.Tags, dataplane.Spec.TagSet()); err != nil {
		return err
	}
	return nil
}

func validateTags(tokenTags v1alpha1.MultiValueTagSet, dpTags v1alpha1.MultiValueTagSet) error {
	if len(tokenTags) != 0 {
		for tagName, allowedValues := range tokenTags {
			dpValues, exist := dpTags[tagName]
			if !exist {
				return errors.Errorf("dataplane has no tag %q required by the token", tagName)
			}
			for value := range dpValues {
				if !allowedValues[value] {
					return errors.Errorf("dataplane contains tag %q with value %q which is not allowed with this token. Allowed values in token are %q", tagName, value, tokenTags.Values(tagName))
				}
			}
		}
	}
	return nil
}
