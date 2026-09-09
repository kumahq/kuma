package labels

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// The zero value is for callers that must not enforce origin and zone, such as the
// admission webhooks, which validate the user-supplied labels instead.
type ControlPlane struct {
	Mode config_core.CpMode
	Zone string
}

func ControlPlaneFromConfig(cfg kuma_cp.Config) ControlPlane {
	return ControlPlane{Mode: cfg.Mode, Zone: cfg.Multizone.Zone.Name}
}

// Descriptor and Spec are taken apart from core_model.Resource because its Meta is
// not set yet when this runs.
type StoredResource struct {
	Descriptor core_model.ResourceTypeDescriptor
	Spec       core_model.ResourceSpec
	Namespace  Namespace
	IsLocal    bool
}

// KDS only writes into the system namespace, so anything outside it is local by
// construction. Inside it, and on Universal, the stored origin is trusted because the
// API server recomputes it on every write and the CP is the only other writer.
func NewStoredResource(res core_model.Resource, ns Namespace, storedLabels map[string]string, cp ControlPlane) StoredResource {
	return StoredResource{
		Descriptor: res.Descriptor(),
		Spec:       res.GetSpec(),
		Namespace:  ns,
		IsLocal:    (ns.value != "" && !ns.system) || core_model.IsLocallyOriginated(cp.Mode, storedLabels),
	}
}

// EnforcedReadLabels recomputes the control-plane-owned labels from the object's own
// identity, never from storage: a stored object can predate a label or have been
// written while the defaulting webhook did not cover its namespace.
func EnforcedReadLabels(r StoredResource, cp ControlPlane) map[string]string {
	enforced := map[string]string{}

	if r.Namespace.value != "" && !r.Namespace.system {
		// The namespace label is always set alongside the role: a workload-owner policy
		// without it matches no dataplane at all, since dppSelectedByNamespace requires
		// the label to be present.
		enforced[mesh_proto.KubeNamespaceTag] = r.Namespace.value
		if r.Descriptor.IsPolicy && r.Descriptor.IsPluginOriginated {
			if policy, ok := r.Spec.(core_model.Policy); ok {
				role, err := ComputePolicyRole(policy, r.Namespace)
				if err != nil {
					// Only reachable for a policy admission never validated. Fall back to the
					// narrowest role instead of erroring: this runs on every read and ToCoreList
					// aborts on the first failure, so one bad object would break matching mesh-wide.
					role = mesh_proto.WorkloadOwnerPolicyRole
				}
				enforced[mesh_proto.PolicyRoleLabel] = string(role)
			}
		}
	}

	switch cp.Mode {
	case config_core.Global:
		if r.IsLocal {
			enforced[mesh_proto.ResourceOriginLabel] = string(mesh_proto.GlobalResourceOrigin)
		} else {
			enforced[mesh_proto.ResourceOriginLabel] = string(mesh_proto.ZoneResourceOrigin)
		}
	case config_core.Zone:
		if r.IsLocal {
			enforced[mesh_proto.ResourceOriginLabel] = string(mesh_proto.ZoneResourceOrigin)
			if cp.Zone != "" && r.Descriptor.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
				enforced[mesh_proto.ZoneTag] = cp.Zone
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
