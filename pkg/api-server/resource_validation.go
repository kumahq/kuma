package api_server

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
)

func (r *resourceCrudHandler) validateResourceRequest(name string, meshName string, resource rest.Resource) error {
	var err validators.ValidationError
	if name != resource.GetMeta().Name {
		err.AddViolation("name", "name from the URL has to be the same as in body")
	}
	if r.federatedZone && !r.doesNameLengthFitsGlobal(name) {
		err.AddViolation("name", "the length of the name must be shorter")
	}
	if string(r.descriptor.Name) != resource.GetMeta().Type {
		err.AddViolation("type", "type from the URL has to be the same as in body")
	}
	if r.descriptor.Scope == core_model.ScopeMesh && meshName != resource.GetMeta().Mesh {
		err.AddViolation("mesh", "mesh from the URL has to be the same as in body")
	}

	err.AddError("labels", r.validateLabels(resource))
	err.AddError("", core_mesh.ValidateMeta(resource.GetMeta(), r.descriptor.Scope))

	return err.OrNil()
}

func (r *resourceCrudHandler) validateOriginForWrite(meta core_model.ResourceMeta) validators.ValidationError {
	var err validators.ValidationError
	origin, ok := core_model.ResourceOrigin(meta)

	if !r.disableOriginLabelValidation && r.mode == config_core.Global {
		if ok && origin != mesh_proto.GlobalResourceOrigin {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ResourceOriginLabel), fmt.Sprintf("the origin label must be set to '%s'", mesh_proto.GlobalResourceOrigin))
		}
	}

	if !r.disableOriginLabelValidation && r.federatedZone && r.descriptor.IsPluginOriginated {
		if !ok || origin != mesh_proto.ZoneResourceOrigin {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ResourceOriginLabel), fmt.Sprintf("the origin label must be set to '%s'", mesh_proto.ZoneResourceOrigin))
		}
	}
	return err
}

func (r *resourceCrudHandler) validateLabels(resource rest.Resource) validators.ValidationError {
	var err validators.ValidationError

	origin, ok := core_model.ResourceOrigin(resource.GetMeta())
	if ok {
		if oerr := origin.IsValid(); oerr != nil {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ResourceOriginLabel), oerr.Error())
		}
	}

	err.AddError("", r.validateOriginForWrite(resource.GetMeta()))

	if r.mode != config_core.Global {
		if origin != mesh_proto.GlobalResourceOrigin {
			zoneTag, ok := resource.GetMeta().GetLabels()[mesh_proto.ZoneTag]
			if ok && zoneTag != r.zoneName {
				err.AddViolationAt(validators.Root().Key(mesh_proto.ZoneTag), fmt.Sprintf("%s label should have %s value", mesh_proto.ZoneTag, r.zoneName))
			}
			if meshLabelValue, ok := resource.GetMeta().GetLabels()[mesh_proto.MeshTag]; ok && meshLabelValue != resource.GetMeta().GetMesh() {
				err.AddViolationAt(validators.Root().Key(mesh_proto.MeshTag), fmt.Sprintf("%s label must not differ from mesh set on resource", mesh_proto.MeshTag))
			}
		}
	}

	if r.descriptor.IsPluginOriginated && r.descriptor.IsPolicy {
		err.AddError("", r.validatePolicyRole(resource))
	}

	for _, k := range maps.SortedKeys(resource.GetMeta().GetLabels()) {
		for _, msg := range validation.IsQualifiedName(k) {
			err.AddViolationAt(validators.Root().Key(k), msg)
		}
		for _, msg := range validation.IsValidLabelValue(resource.GetMeta().GetLabels()[k]) {
			err.AddViolationAt(validators.Root().Key(k), msg)
		}
	}
	return err
}

func (r *resourceCrudHandler) validateImmutableLabels(currentComputedLabels, newComputedLabels map[string]string) validators.ValidationError {
	var err validators.ValidationError

	immutableLabels := []string{
		mesh_proto.ZoneTag,
	}

	for _, label := range immutableLabels {
		currentVal, currentExists := currentComputedLabels[label]
		newVal, newExists := newComputedLabels[label]

		if currentExists && !newExists {
			err.AddViolationAt(
				validators.Root().Key(label),
				fmt.Sprintf("is immutable, cannot be removed (was %q)", currentVal),
			)
		} else if currentExists && currentVal != newVal {
			err.AddViolationAt(
				validators.Root().Key(label),
				fmt.Sprintf("is immutable, cannot be changed from %q to %q", currentVal, newVal),
			)
		}
	}

	return err
}

func (r *resourceCrudHandler) validatePolicyRole(resource rest.Resource) validators.ValidationError {
	var err validators.ValidationError
	policyRole := core_model.PolicyRole(resource.GetMeta())
	// at the moment on universal all policies have system policy role
	if policyRole != mesh_proto.SystemPolicyRole {
		err.AddViolationAt(validators.Root().Key(mesh_proto.PolicyRoleLabel), fmt.Sprintf("%s label should have %s value, got %s", mesh_proto.PolicyRoleLabel, mesh_proto.SystemPolicyRole, policyRole))
	}
	return err
}

// The resource is prefixed with the zone name when it is synchronized
// to global control-plane. It is important to notice that the zone is unaware
// of the type of the store used by the global control-plane, so we must prepare
// for the worst-case scenario. We don't have to check other plugabble policies
// because zone doesn't allow to create policies on the zone.
func (r *resourceCrudHandler) doesNameLengthFitsGlobal(name string) bool {
	return len(fmt.Sprintf("%s.%s", r.zoneName, name)) < 253
}
