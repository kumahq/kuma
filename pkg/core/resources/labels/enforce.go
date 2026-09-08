package labels

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// EnforcedReadLabels returns the control-plane-owned labels recomputed from the
// object's identity rather than read back from storage: the owning zone, the
// namespace, and the policy role. Returns nil when there is nothing to enforce.
//
// Namespace and role are skipped for UnsetNamespace (Universal, KDS metas) and
// the system namespace, where stored values are mesh-wide or KDS imports.
func EnforcedReadLabels(
	rd core_model.ResourceTypeDescriptor,
	spec core_model.ResourceSpec,
	stored map[string]string,
	opts ...Option,
) map[string]string {
	o := NewOptions(opts...)
	enforced := map[string]string{}

	// After a zone rename stored policies keep the old kuma.io/zone while
	// Dataplanes are rewritten, so dppSelectedByZone drops them. Origin is
	// trusted: KDS stamps global on every import, Compute stamps zone on every
	// local write, and matching ignores the zone label when origin is absent.
	if o.Mode == config_core.Zone && rd.KDSFlags.Has(core_model.ProvidedByZoneFlag) &&
		stored[mesh_proto.ResourceOriginLabel] == string(mesh_proto.ZoneResourceOrigin) {
		enforced[mesh_proto.ZoneTag] = o.ZoneName
	}

	ns := o.Namespace
	if ns.value != "" && !ns.system {
		// The namespace label is always set alongside the role: a workload-owner policy
		// without it matches no dataplane at all, since dppSelectedByNamespace requires
		// the label to be present.
		enforced[mesh_proto.KubeNamespaceTag] = ns.value
		if rd.IsPolicy && rd.IsPluginOriginated {
			if policy, ok := spec.(core_model.Policy); ok {
				role, err := ComputePolicyRole(policy, ns)
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

	if len(enforced) == 0 {
		return nil
	}
	return enforced
}
