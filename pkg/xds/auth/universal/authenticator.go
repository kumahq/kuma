package universal

import (
	"context"
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	builtin_issuer "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/xds/auth"
)

func NewAuthenticator(dataplaneValidator builtin_issuer.Validator, zoneValidator zone.Validator, zone string) auth.Authenticator {
	return &universalAuthenticator{
		dataplaneValidator: dataplaneValidator,
		zoneValidator:      zoneValidator,
		zone:               zone,
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
	dataplaneValidator builtin_issuer.Validator
	zoneValidator      zone.Validator
	zone               string
}

var _ auth.Authenticator = &universalAuthenticator{}

func (u *universalAuthenticator) Authenticate(ctx context.Context, resource model.Resource, credential auth.Credential) error {
	switch resource := resource.(type) {
	case *core_mesh.DataplaneResource:
		return u.authDataplane(ctx, resource, credential)
	case *core_mesh.ZoneIngressResource:
		return u.authZoneEntity(ctx, credential, zone.IngressScope)
	case *core_mesh.ZoneEgressResource:
		return u.authZoneEntity(ctx, credential, zone.EgressScope)
	default:
		return fmt.Errorf("no matching authenticator for %s resource", resource.Descriptor().Name)
	}
}

func (u *universalAuthenticator) authDataplane(ctx context.Context, dataplane *core_mesh.DataplaneResource, credential auth.Credential) error {
	dpIdentity, err := u.dataplaneValidator.Validate(ctx, credential, dataplane.Meta.GetMesh())
	if err != nil {
		return err
	}

	if dpIdentity.Name != "" && dataplane.Meta.GetName() != dpIdentity.Name {
		return fmt.Errorf("proxy name from requestor: %s is different than in token: %s", dataplane.Meta.GetName(), dpIdentity.Name)
	}
	if dpIdentity.Mesh != "" && dataplane.Meta.GetMesh() != dpIdentity.Mesh {
		return fmt.Errorf("proxy mesh from requestor: %s is different than in token: %s", dataplane.Meta.GetMesh(), dpIdentity.Mesh)
	}
	if err := validateTags(dpIdentity.Tags, dataplane.Spec.TagSet()); err != nil {
		return err
	}
	return nil
}

func (u *universalAuthenticator) authZoneEntity(
	ctx context.Context,
	credential auth.Credential,
	scope string,
) error {
	identity, err := u.zoneValidator.Validate(ctx, credential)
	if err != nil {
		return err
	}

	if !zone.InScope(identity.Scope, scope) {
		return fmt.Errorf(
			"token cannot be used to authenticate zone entity (%s is out of token's scope: %+v)",
			scope,
			identity.Scope,
		)
	}

	if identity.Zone != "" && u.zone != identity.Zone {
		return fmt.Errorf("zone from requestor: %s is different than in token: %s", u.zone, identity.Zone)
	}

	return nil
}

func validateTags(tokenTags mesh_proto.MultiValueTagSet, dpTags mesh_proto.MultiValueTagSet) error {
	for tagName, allowedValues := range tokenTags {
		dpValues, exist := dpTags[tagName]
		if !exist {
			return fmt.Errorf("dataplane has no tag %q required by the token", tagName)
		}
		for value := range dpValues {
			if !allowedValues[value] {
				return fmt.Errorf("dataplane contains tag %q with value %q which is not allowed with this token. Allowed values in token are %q", tagName, value, tokenTags.Values(tagName))
			}
		}
	}
	return nil
}
