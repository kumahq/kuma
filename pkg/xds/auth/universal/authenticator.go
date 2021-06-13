package universal

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	builtin_issuer "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/xds/auth"
)

func NewAuthenticator(issuer builtin_issuer.DataplaneTokenIssuer, zoneIngressIssuer zoneingress.TokenIssuer, zone string) auth.Authenticator {
	return &universalAuthenticator{
		issuer:            issuer,
		zoneIngressIssuer: zoneIngressIssuer,
		zone:              zone,
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
	zoneIngressIssuer zoneingress.TokenIssuer
	zone              string
}

func (u *universalAuthenticator) Authenticate(ctx context.Context, dataplane *core_mesh.DataplaneResource, credential auth.Credential) error {
	dpIdentity, err := u.issuer.Validate(credential, dataplane.Meta.GetMesh())
	if err != nil {
		return err
	}

	if dpIdentity.Name != "" && dataplane.Meta.GetName() != dpIdentity.Name {
		return errors.Errorf("proxy name from requestor: %s is different than in token: %s", dataplane.Meta.GetName(), dpIdentity.Name)
	}
	if dpIdentity.Mesh != "" && dataplane.Meta.GetMesh() != dpIdentity.Mesh {
		return errors.Errorf("proxy mesh from requestor: %s is different than in token: %s", dataplane.Meta.GetMesh(), dpIdentity.Mesh)
	}
	if err := validateType(dataplane, dpIdentity.Type); err != nil {
		return err
	}
	if err := validateTags(dpIdentity.Tags, dataplane.Spec.TagSet()); err != nil {
		return err
	}
	return nil
}

func (u *universalAuthenticator) AuthenticateZoneIngress(ctx context.Context, zoneIngress *core_mesh.ZoneIngressResource, credential auth.Credential) error {
	identity, err := u.zoneIngressIssuer.Validate(credential)
	if err != nil {
		return err
	}
	if u.zone != identity.Zone {
		return errors.Errorf("zone ingress zone from requestor: %s is different than in token: %s", u.zone, identity.Zone)
	}

	return nil
}

func validateType(dataplane *core_mesh.DataplaneResource, proxyType mesh_proto.ProxyType) error {
	if proxyType == "" { // if dp type is not explicitly specified  we assume it's dataplane so we force Ingress token
		proxyType = mesh_proto.DataplaneProxyType
	}
	if dataplane.Spec.IsIngress() && proxyType != mesh_proto.IngressProxyType {
		return errors.Errorf("dataplane is of type Ingress but token allows only for the %q type", proxyType)
	}
	if !dataplane.Spec.IsIngress() && proxyType == mesh_proto.IngressProxyType {
		return errors.Errorf("dataplane is of type Dataplane but token allows only for the %q type", proxyType)
	}
	return nil
}

func validateTags(tokenTags mesh_proto.MultiValueTagSet, dpTags mesh_proto.MultiValueTagSet) error {
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
	return nil
}
