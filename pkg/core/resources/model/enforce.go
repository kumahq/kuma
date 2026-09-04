package model

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
)

// EnforcedReadLabels returns the control-plane-owned labels, recomputed from the
// object's own identity instead of being read back from storage: the namespace it
// really lives in, and the policy role that follows from that namespace and the spec.
// A stored object can disagree with either - it predates the label, or it was written
// while the defaulting webhook did not cover its namespace.
//
// Returns nil when there is nothing to enforce: UnsetNamespace (Universal, and KDS
// metas, which carry no name extensions) and the system namespace, where the stored
// values are either legitimately mesh-wide or KDS imports whose labels record the
// producing CP's decision.
func EnforcedReadLabels(
	rd ResourceTypeDescriptor,
	spec ResourceSpec,
	ns Namespace,
) map[string]string {
	if ns.value == "" || ns.system {
		return nil
	}
	// The namespace label is always set alongside the role: a workload-owner policy
	// without it matches no dataplane at all, since dppSelectedByNamespace requires
	// the label to be present.
	enforced := map[string]string{mesh_proto.KubeNamespaceTag: ns.value}
	if rd.IsPolicy && rd.IsPluginOriginated {
		if policy, ok := spec.(Policy); ok {
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
	return enforced
}
