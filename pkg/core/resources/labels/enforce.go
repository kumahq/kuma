package labels

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// ControlPlane is the configuration fixed for the whole control plane deployment.
// The zero value belongs to a caller that must not enforce origin and zone, such as
// the admission webhooks, which validate the user-supplied labels instead.
type ControlPlane struct {
	Mode config_core.CpMode
	Zone string
}

func ControlPlaneFromConfig(cfg kuma_cp.Config) ControlPlane {
	return ControlPlane{Mode: cfg.Mode, Zone: cfg.Multizone.Zone.Name}
}

// StoredResource is the object whose labels are recomputed, as it came out of storage.
// Descriptor and Spec are taken separately rather than as a core_model.Resource because
// the resource's Meta is not set yet when this runs.
type StoredResource struct {
	Descriptor core_model.ResourceTypeDescriptor
	Spec       core_model.ResourceSpec
	Namespace  Namespace
	// IsLocal is true when this control plane authored the object rather than
	// received it over KDS.
	IsLocal bool
}

// NewStoredResource derives IsLocal with one rule for both stores: KDS only ever writes
// into the system namespace, so an object outside it is local by construction; inside it,
// and on Universal, the stored origin is the only signal and is trusted because the API
// server recomputes it on every write and the CP is the only other writer.
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
