package labels

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// EnforcedReadLabels recomputes the control-plane-owned labels from the object's own
// identity, never from storage: a stored object can predate a label or have been
// written while the defaulting webhook did not cover its namespace.
func EnforcedReadLabels(
	rd core_model.ResourceTypeDescriptor,
	spec core_model.ResourceSpec,
	isLocal bool,
	opts ...Option,
) map[string]string {
	o := NewOptions(opts...)
	enforced := map[string]string{}

	if o.Namespace.value != "" && !o.Namespace.system {
		// The namespace label is always set alongside the role: a workload-owner policy
		// without it matches no dataplane at all, since dppSelectedByNamespace requires
		// the label to be present.
		enforced[mesh_proto.KubeNamespaceTag] = o.Namespace.value
		if rd.IsPolicy && rd.IsPluginOriginated {
			if policy, ok := spec.(core_model.Policy); ok {
				role, err := ComputePolicyRole(policy, o.Namespace)
				if err != nil {
					// Only reachable for a policy admission never validated (mixed
					// producer and consumer items). Fall back to the narrowest role
					// rather than returning an error: this runs on every read, and
					// ToCoreList aborts on the first failure, so one malformed stored
					// object would otherwise break policy matching mesh-wide.
					role = mesh_proto.WorkloadOwnerPolicyRole
				}
				enforced[mesh_proto.PolicyRoleLabel] = string(role)
			}
		}
	}

	switch o.Mode {
	case config_core.Global:
		if isLocal {
			enforced[mesh_proto.ResourceOriginLabel] = string(mesh_proto.GlobalResourceOrigin)
		} else {
			enforced[mesh_proto.ResourceOriginLabel] = string(mesh_proto.ZoneResourceOrigin)
		}
	case config_core.Zone:
		if isLocal {
			enforced[mesh_proto.ResourceOriginLabel] = string(mesh_proto.ZoneResourceOrigin)
			if o.ZoneName != "" && rd.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
				enforced[mesh_proto.ZoneTag] = o.ZoneName
			}
		} else {
			enforced[mesh_proto.ResourceOriginLabel] = string(mesh_proto.GlobalResourceOrigin)
		}
	}

	if len(enforced) == 0 {
		return nil
	}
	return enforced
}
