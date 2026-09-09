package labels

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
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
	rd core_model.ResourceTypeDescriptor,
	spec core_model.ResourceSpec,
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
	return enforced
}

// EnforcedZoneLabel returns the kuma.io/zone value a stored resource must be read
// with, or "" when the label is not this control plane's to own. Only a zone
// control plane owns it: on Global the stored value records the producing zone's
// decision and has to survive the read.
//
// kuma.io/zone is written once, at admission time, and never revisited. Renaming a
// zone - which is what federating a standalone zone does - therefore leaves every
// resource written beforehand claiming the old name while everything written after
// claims the new one. Policy matching reads the label back to decide whether a
// zone-origin policy belongs to the proxy's zone, so the mismatch silently drops
// every pre-rename policy from every proxy.
func EnforcedZoneLabel(
	rd core_model.ResourceTypeDescriptor,
	mode config_core.CpMode,
	zone string,
	stored map[string]string,
) string {
	if mode != config_core.Zone || zone == "" {
		return ""
	}
	if !rd.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
		return ""
	}
	if !core_model.IsLocallyOriginated(mode, stored) {
		return ""
	}
	return zone
}
