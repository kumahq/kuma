package universal

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	builtin_issuer "github.com/kumahq/kuma/v3/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/v3/pkg/xds/auth"
)

func NewAuthenticator(
	dataplaneValidator builtin_issuer.Validator,
	zoneValidator zone.Validator,
	resManager manager.ReadOnlyResourceManager,
	env config_core.EnvironmentType,
	zone string,
) auth.Authenticator {
	return &universalAuthenticator{
		dataplaneValidator: dataplaneValidator,
		zoneValidator:      zoneValidator,
		resManager:         resManager,
		env:                env,
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
	resManager         manager.ReadOnlyResourceManager
	env                config_core.EnvironmentType
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
		return errors.Errorf("no matching authenticator for %s resource", resource.Descriptor().Name)
	}
}

func (u *universalAuthenticator) authDataplane(ctx context.Context, dataplane *core_mesh.DataplaneResource, credential auth.Credential) error {
	dpIdentity, err := u.dataplaneValidator.Validate(ctx, credential, dataplane.Meta.GetMesh())
	if err != nil {
		return err
	}

	if dpIdentity.Name != "" && dataplane.Meta.GetName() != dpIdentity.Name {
		return errors.Errorf("proxy name from requestor: %s is different than in token: %s", dataplane.Meta.GetName(), dpIdentity.Name)
	}
	if dpIdentity.Mesh != "" && dataplane.Meta.GetMesh() != dpIdentity.Mesh {
		return errors.Errorf("proxy mesh from requestor: %s is different than in token: %s", dataplane.Meta.GetMesh(), dpIdentity.Mesh)
	}
	if err := validateTags(dpIdentity.Tags, dataplane.Spec.TagSet()); err != nil {
		return err
	}
	identityFromWorkload, err := u.identityDerivesFromWorkloadLabel(ctx, dataplane)
	if err != nil {
		return err
	}
	if err := validateWorkload(dpIdentity.Workload, dataplane.Meta.GetLabels(), identityFromWorkload); err != nil {
		return err
	}
	return nil
}

// identityDerivesFromWorkloadLabel reports whether the dataplane's SPIFFE
// identity is derived from its kuma.io/workload label. This is the case when a
// MeshIdentity selects the dataplane and its SPIFFE ID path template references
// the workload (the default in universal mode). When true, the workload label
// becomes the proxy's mTLS identity, so the presented token must be bound to
// that workload (see validateWorkload).
func (u *universalAuthenticator) identityDerivesFromWorkloadLabel(ctx context.Context, dataplane *core_mesh.DataplaneResource) (bool, error) {
	meshIdentities := &meshidentity_api.MeshIdentityResourceList{}
	if err := u.resManager.List(ctx, meshIdentities, store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return false, errors.Wrap(err, "failed to list MeshIdentities")
	}
	matched, found := meshidentity_api.BestMatched(dataplane.Meta.GetLabels(), meshIdentities.Items)
	if !found {
		return false, nil
	}
	return matched.Spec.UsesWorkloadLabel(u.env), nil
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
		return errors.Errorf(
			"token cannot be used to authenticate zone entity (%s is out of token's scope: %+v)",
			scope,
			identity.Scope,
		)
	}

	if identity.Zone != "" && u.zone != identity.Zone {
		return errors.Errorf("zone from requestor: %s is different than in token: %s", u.zone, identity.Zone)
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

func validateWorkload(tokenWorkload string, dpLabels map[string]string, identityFromWorkloadLabel bool) error {
	dpWorkload, exists := dpLabels[metadata.KumaWorkload]
	if tokenWorkload == "" {
		// When the dataplane's SPIFFE identity is derived from the kuma.io/workload
		// label, that label becomes the proxy's identity. A token that is not bound
		// to a workload does not constrain the label, so accepting it would let the
		// dataplane pick an arbitrary workload identity. Require a workload-bound
		// token so the token's scope stays consistent with the issued identity.
		if identityFromWorkloadLabel && exists {
			return errors.Errorf("dataplane with %q label %q derives its identity from the workload, so it requires a workload-bound token", metadata.KumaWorkload, dpWorkload)
		}
		return nil
	}
	if !exists {
		return errors.Errorf("dataplane has no workload label required by the token")
	}
	if dpWorkload != tokenWorkload {
		return errors.Errorf("dataplane workload %q is not allowed with this token. Allowed workload in token is %q", dpWorkload, tokenWorkload)
	}
	return nil
}
