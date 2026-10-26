package universal

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	builtin_issuer "github.com/kumahq/kuma/v3/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/v3/pkg/xds/auth"
)

func NewAuthenticator(
	dataplaneValidator builtin_issuer.Validator,
	resManager manager.ReadOnlyResourceManager,
	connectionInfoTracker core_xds.ConnectionInfoTracker,
) auth.Authenticator {
	return &universalAuthenticator{
		dataplaneValidator:    dataplaneValidator,
		resManager:            resManager,
		connectionInfoTracker: connectionInfoTracker,
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
type universalAuthenticator struct {
	dataplaneValidator    builtin_issuer.Validator
	resManager            manager.ReadOnlyResourceManager
	connectionInfoTracker core_xds.ConnectionInfoTracker
}

var _ auth.Authenticator = &universalAuthenticator{}

func (u *universalAuthenticator) Authenticate(ctx context.Context, resource model.Resource, credential auth.Credential) error {
	switch resource := resource.(type) {
	case *core_mesh.DataplaneResource:
		return u.authDataplane(ctx, resource, credential)
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
	if err := validateTags(dpIdentity.Tags, dataplane); err != nil {
		return err
	}
	if err := u.validateWorkload(ctx, dpIdentity.Workload, dataplane); err != nil {
		return err
	}
	// Persist the binding so issuance can re-check a MeshIdentity added post-connect.
	if u.connectionInfoTracker != nil {
		u.connectionInfoTracker.Store(model.MetaToResourceKey(dataplane.GetMeta()), core_xds.ConnectionInfo{TokenWorkload: dpIdentity.Workload})
	}
	return nil
}

// Whether a SPIFFE ID template reads kuma.io/workload is not decidable by inspecting
// it, so the label counts as identity whenever any MeshIdentity selects the dataplane.
func (u *universalAuthenticator) selectedByMeshIdentity(ctx context.Context, dataplane *core_mesh.DataplaneResource) (bool, error) {
	meshIdentities := &meshidentity_api.MeshIdentityResourceList{}
	if err := u.resManager.List(ctx, meshIdentities, store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return false, errors.Wrap(err, "failed to list MeshIdentities")
	}
	_, found := meshidentity_api.BestMatched(dataplane.Meta.GetLabels(), meshIdentities.Items)
	return found, nil
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

// validateWorkload checks the token's workload binding against the dataplane's
// kuma.io/workload label. The MeshIdentity lookup runs lazily, only when it can
// change the result: an unbound token on a dataplane that declares that label.
func (u *universalAuthenticator) validateWorkload(ctx context.Context, tokenWorkload string, dataplane *core_mesh.DataplaneResource) error {
	dpWorkload, exists := dataplane.Meta.GetLabels()[metadata.KumaWorkload]
	if tokenWorkload == "" {
		if !exists {
			return nil
		}
		// An unbound token can't constrain a label that may be the proxy's identity.
		selected, err := u.selectedByMeshIdentity(ctx, dataplane)
		if err != nil {
			return err
		}
		if selected {
			return errors.Errorf("dataplane with %q label %q is selected by a MeshIdentity, so it requires a workload-bound token", metadata.KumaWorkload, dpWorkload)
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
