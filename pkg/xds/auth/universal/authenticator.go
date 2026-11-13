package universal

import (
	"context"
	"maps"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	builtin_issuer "github.com/kumahq/kuma/v2/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/v2/pkg/xds/auth"
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
//
// A Dataplane selected by a MeshIdentity additionally needs every label its SPIFFE ID
// is rendered from to be bound by the token, via the workload or a tag.
type universalAuthenticator struct {
	dataplaneValidator builtin_issuer.Validator
	zoneValidator      zone.Validator
	resManager         manager.ReadOnlyResourceManager
	env                config_core.EnvironmentType
	zone               string
}

var _ auth.Authenticator = &universalAuthenticator{}

func (u *universalAuthenticator) Authenticate(ctx context.Context, resource model.Resource, credential auth.Credential) (auth.Binding, error) {
	switch resource := resource.(type) {
	case *core_mesh.DataplaneResource:
		return u.authDataplane(ctx, resource, credential)
	case *core_mesh.ZoneIngressResource:
		return u.authZoneEntity(ctx, credential, zone.IngressScope)
	case *core_mesh.ZoneEgressResource:
		return u.authZoneEntity(ctx, credential, zone.EgressScope)
	default:
		return auth.Binding{}, errors.Errorf("no matching authenticator for %s resource", resource.Descriptor().Name)
	}
}

func (u *universalAuthenticator) authDataplane(ctx context.Context, dataplane *core_mesh.DataplaneResource, credential auth.Credential) (auth.Binding, error) {
	dpIdentity, err := u.dataplaneValidator.Validate(ctx, credential, dataplane.Meta.GetMesh())
	if err != nil {
		return auth.Binding{}, err
	}

	if dpIdentity.Name != "" && dataplane.Meta.GetName() != dpIdentity.Name {
		return auth.Binding{}, errors.Errorf("proxy name from requestor: %s is different than in token: %s", dataplane.Meta.GetName(), dpIdentity.Name)
	}
	if dpIdentity.Mesh != "" && dataplane.Meta.GetMesh() != dpIdentity.Mesh {
		return auth.Binding{}, errors.Errorf("proxy mesh from requestor: %s is different than in token: %s", dataplane.Meta.GetMesh(), dpIdentity.Mesh)
	}
	if err := validateTags(dpIdentity.Tags, dataplane.Spec.TagSet()); err != nil {
		return auth.Binding{}, err
	}
	if err := validateWorkload(dpIdentity.Workload, dataplane); err != nil {
		return auth.Binding{}, err
	}
	binding := auth.Binding{Labels: provenLabels(dpIdentity.Tags, dpIdentity.Workload, dpIdentity.Name, dataplane.Meta.GetMesh())}
	if err := u.validateIdentityLabels(ctx, binding, dataplane); err != nil {
		return auth.Binding{}, err
	}
	return binding, nil
}

// provenLabels merges the token's workload and name claims into its tags. When
// a claim and a tag constrain the same label, only values both allow are proven.
// On Universal kuma.io/display-name is the dataplane name, so the name claim proves it.
// The token is validated with the key of mesh, so it proves kuma.io/mesh too.
func provenLabels(tags mesh_proto.MultiValueTagSet, workload, name, mesh string) mesh_proto.MultiValueTagSet {
	proven := mesh_proto.MultiValueTagSet{}
	maps.Copy(proven, tags)
	for key, claim := range map[string]string{metadata.KumaWorkload: workload, mesh_proto.DisplayName: name, metadata.KumaMeshLabel: mesh} {
		if claim == "" {
			continue
		}
		tagged, ok := proven[key]
		proven[key] = map[string]bool{claim: !ok || tagged[claim]}
	}
	return proven
}

// validateIdentityLabels rejects at connect what identity issuance refuses
// anyway, so the proxy fails loudly instead of running without an identity.
func (u *universalAuthenticator) validateIdentityLabels(ctx context.Context, binding auth.Binding, dataplane *core_mesh.DataplaneResource) error {
	meshIdentities := &meshidentity_api.MeshIdentityResourceList{}
	if err := u.resManager.List(ctx, meshIdentities, store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return errors.Wrap(err, "failed to list MeshIdentities")
	}
	matched, found := meshidentity_api.BestMatched(dataplane.Meta.GetLabels(), meshIdentities.Items)
	if !found {
		return nil
	}
	// A render error is left to issuance, which renders the same inputs and fails too.
	_, read, _ := matched.RenderSpiffeID(dataplane.GetMeta(), u.zone, u.env)
	return auth.ValidateIdentityLabels(binding, dataplane.Meta.GetLabels(), read)
}

func (u *universalAuthenticator) authZoneEntity(
	ctx context.Context,
	credential auth.Credential,
	scope string,
) (auth.Binding, error) {
	identity, err := u.zoneValidator.Validate(ctx, credential)
	if err != nil {
		return auth.Binding{}, err
	}

	if !zone.InScope(identity.Scope, scope) {
		return auth.Binding{}, errors.Errorf(
			"token cannot be used to authenticate zone entity (%s is out of token's scope: %+v)",
			scope,
			identity.Scope,
		)
	}

	if identity.Zone != "" && u.zone != identity.Zone {
		return auth.Binding{}, errors.Errorf("zone from requestor: %s is different than in token: %s", u.zone, identity.Zone)
	}

	return auth.Binding{}, nil
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

// validateWorkload checks a workload-bound token against the dataplane's
// kuma.io/workload label.
func validateWorkload(tokenWorkload string, dataplane *core_mesh.DataplaneResource) error {
	if tokenWorkload == "" {
		return nil
	}
	dpWorkload, exists := dataplane.Meta.GetLabels()[metadata.KumaWorkload]
	if !exists {
		return errors.Errorf("dataplane has no workload label required by the token")
	}
	if dpWorkload != tokenWorkload {
		return errors.Errorf("dataplane workload %q is not allowed with this token. Allowed workload in token is %q", dpWorkload, tokenWorkload)
	}
	return nil
}
