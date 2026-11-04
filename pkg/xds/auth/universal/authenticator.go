package universal

import (
	"context"
	"maps"

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
	"github.com/kumahq/kuma/v3/pkg/xds/auth"
)

func NewAuthenticator(
	dataplaneValidator builtin_issuer.Validator,
	resManager manager.ReadOnlyResourceManager,
	zone string,
	env config_core.EnvironmentType,
) auth.Authenticator {
	return &universalAuthenticator{
		dataplaneValidator: dataplaneValidator,
		resManager:         resManager,
		zone:               zone,
		env:                env,
	}
}

// universalAuthenticator defines authentication for Dataplane Tokens
// All fields in token are optional, so we only validate data that is available in token. This way you can pick your level of security.
// Generate token for mesh+name for maximum security.
// Generate token for mesh+tags (matched against Dataplane labels, e.g. kuma.io/workload)
// so you can reuse the token for many dataplanes.
// Generate token for mesh if you trust the scope of the mesh.
//
// If you generate token bound to tags, the Dataplane needs to have all the tags defined in
// the token, and every value it declares for a tag has to be allowed by the token.
//
// A Dataplane selected by a MeshIdentity additionally needs every label its SPIFFE ID
// is rendered from to be bound by the token, via the workload or a tag.
type universalAuthenticator struct {
	dataplaneValidator builtin_issuer.Validator
	resManager         manager.ReadOnlyResourceManager
	zone               string
	env                config_core.EnvironmentType
}

var _ auth.Authenticator = &universalAuthenticator{}

func (u *universalAuthenticator) Authenticate(ctx context.Context, resource model.Resource, credential auth.Credential) (auth.Binding, error) {
	switch resource := resource.(type) {
	case *core_mesh.DataplaneResource:
		return u.authDataplane(ctx, resource, credential)
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
	if err := validateTags(dpIdentity.Tags, dataplane); err != nil {
		return auth.Binding{}, err
	}
	if err := validateWorkload(dpIdentity.Workload, dataplane); err != nil {
		return auth.Binding{}, err
	}
	binding := auth.Binding{Labels: provenLabels(dpIdentity.Tags, dpIdentity.Workload)}
	if err := u.validateIdentityLabels(ctx, binding, dataplane); err != nil {
		return auth.Binding{}, err
	}
	return binding, nil
}

// provenLabels merges the token's workload claim into its tags. When both
// constrain kuma.io/workload, only values both allow are proven.
func provenLabels(tags mesh_proto.MultiValueTagSet, workload string) mesh_proto.MultiValueTagSet {
	proven := mesh_proto.MultiValueTagSet{}
	maps.Copy(proven, tags)
	if workload != "" {
		tagged, ok := proven[metadata.KumaWorkload]
		proven[metadata.KumaWorkload] = map[string]bool{workload: !ok || tagged[workload]}
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

// validateTags checks a tag-bound token against the dataplane's labels.
func validateTags(tokenTags mesh_proto.MultiValueTagSet, dataplane *core_mesh.DataplaneResource) error {
	dpLabels := dataplane.Meta.GetLabels()

	for tagName, allowedValues := range tokenTags {
		dpValues := map[string]bool{}
		if labelValue, exist := dpLabels[tagName]; exist {
			dpValues[labelValue] = true
		}
		if len(dpValues) == 0 {
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
