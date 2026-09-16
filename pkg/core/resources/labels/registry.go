package labels

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
)

type Owner int

const (
	// OwnerControlPlane: the control plane decides the value on every write. An
	// untrusted writer may supply it only with the value Compute would store; any
	// other value is rejected by ValidateOwnership.
	OwnerControlPlane Owner = iota
	// OwnerUser: the user sets the value; only ValidateFormat applies.
	OwnerUser
)

type Descriptor struct {
	Key   string
	Owner Owner

	// Compute returns the value to store on a write; ok=false removes the label.
	// err aborts the whole write.
	Compute func(w Write, cp ControlPlane) (value string, ok bool, err error)

	// EnforceOnRead recomputes the label on every read. ok=false means "produce
	// nothing", never "remove".
	EnforceOnRead func(r StoredResource, cp ControlPlane) (value string, ok bool)

	// ValidateFormat checks the value's syntax for any writer.
	ValidateFormat func(value string) []string

	// StoredAsAnnotation marks labels whose values carry a resource name and therefore
	// can be up to 253 characters, which does not fit the 63-character Kubernetes label
	// value limit. They are stored as annotations on Kubernetes and validated as names.
	StoredAsAnnotation bool
}

func listenerLabel(listenerType mesh_proto.Dataplane_Networking_Listener_Type) func(w Write, cp ControlPlane) (string, bool, error) {
	return func(w Write, _ ControlPlane) (string, bool, error) {
		dp, ok := w.Spec.(*mesh_proto.Dataplane)
		if !w.Descriptor.IsProxy || !ok {
			return "", false, nil
		}
		for _, l := range dp.GetNetworking().GetListeners() {
			if l.Type == listenerType {
				return "enabled", true, nil
			}
		}
		return "", false, nil
	}
}

// Iteration order is the order the API server reports ownership violations in.
var registry = []Descriptor{
	{
		Key:   mesh_proto.ResourceOriginLabel,
		Owner: OwnerControlPlane,
		Compute: func(_ Write, cp ControlPlane) (string, bool, error) {
			switch cp.Mode {
			case config_core.Global:
				return string(mesh_proto.GlobalResourceOrigin), true, nil
			case config_core.Zone:
				return string(mesh_proto.ZoneResourceOrigin), true, nil
			default:
				return "", false, nil
			}
		},
		EnforceOnRead: func(r StoredResource, cp ControlPlane) (string, bool) {
			switch cp.Mode {
			case config_core.Global:
				if r.IsLocal {
					return string(mesh_proto.GlobalResourceOrigin), true
				}
				return string(mesh_proto.ZoneResourceOrigin), true
			case config_core.Zone:
				if r.IsLocal {
					return string(mesh_proto.ZoneResourceOrigin), true
				}
				return string(mesh_proto.GlobalResourceOrigin), true
			default:
				return "", false
			}
		},
	},
	{
		Key:   mesh_proto.ZoneTag,
		Owner: OwnerControlPlane,
		Compute: func(w Write, cp ControlPlane) (string, bool, error) {
			if cp.Mode == config_core.Zone && w.Descriptor.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
				return cp.Zone, true, nil
			}
			return "", false, nil
		},
		EnforceOnRead: func(r StoredResource, cp ControlPlane) (string, bool) {
			if cp.Mode == config_core.Zone && r.IsLocal && cp.Zone != "" && r.Descriptor.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
				return cp.Zone, true
			}
			return "", false
		},
	},
	{
		Key:   metadata.KumaMeshLabel,
		Owner: OwnerControlPlane,
		Compute: func(w Write, _ ControlPlane) (string, bool, error) {
			if w.Descriptor.Scope != core_model.ScopeMesh {
				return "", false, nil
			}
			if w.Mesh != "" {
				return w.Mesh, true, nil
			}
			return core_model.DefaultMesh, true, nil
		},
	},
	{
		Key:   mesh_proto.PolicyRoleLabel,
		Owner: OwnerControlPlane,
		Compute: func(w Write, _ ControlPlane) (string, bool, error) {
			policy, ok := w.Spec.(core_model.Policy)
			if !w.Descriptor.IsPolicy || !w.Descriptor.IsPluginOriginated || !ok {
				return "", false, nil
			}
			role, err := ComputePolicyRole(policy, w.Namespace)
			if err != nil {
				return "", false, err
			}
			return string(role), true, nil
		},
		EnforceOnRead: func(r StoredResource, _ ControlPlane) (string, bool) {
			if r.Namespace.value == "" || r.Namespace.system || !r.Descriptor.IsPolicy || !r.Descriptor.IsPluginOriginated {
				return "", false
			}
			policy, ok := r.Spec.(core_model.Policy)
			if !ok {
				return "", false
			}
			role, err := ComputePolicyRole(policy, r.Namespace)
			if err != nil {
				// Only reachable for a policy admission never validated. Fall back to the
				// narrowest role instead of erroring: this runs on every read and ToCoreList
				// aborts on the first failure, so one bad object would break matching mesh-wide.
				role = mesh_proto.WorkloadOwnerPolicyRole
			}
			return string(role), true
		},
	},
	{
		Key:   mesh_proto.DisplayName,
		Owner: OwnerControlPlane,
		Compute: func(w Write, _ ControlPlane) (string, bool, error) {
			return w.DisplayName, true, nil
		},
		StoredAsAnnotation: true,
	},
	{
		Key:   mesh_proto.EnvTag,
		Owner: OwnerControlPlane,
		Compute: func(w Write, cp ControlPlane) (string, bool, error) {
			if cp.Mode != config_core.Zone || !w.Descriptor.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
				return "", false, nil
			}
			if cp.IsK8s {
				return mesh_proto.KubernetesEnvironment, true, nil
			}
			return mesh_proto.UniversalEnvironment, true, nil
		},
	},
	{
		Key:   mesh_proto.KubeNamespaceTag,
		Owner: OwnerControlPlane,
		Compute: func(w Write, cp ControlPlane) (string, bool, error) {
			if !cp.IsK8s || w.Namespace.value == "" {
				return "", false, nil
			}
			return w.Namespace.value, true, nil
		},
		EnforceOnRead: func(r StoredResource, _ ControlPlane) (string, bool) {
			if r.Namespace.value != "" && !r.Namespace.system {
				return r.Namespace.value, true
			}
			return "", false
		},
	},
	{
		Key:   metadata.KumaServiceAccount,
		Owner: OwnerControlPlane,
		Compute: func(w Write, cp ControlPlane) (string, bool, error) {
			if !cp.IsK8s {
				return "", false, nil
			}
			if w.ServiceAccount != "" {
				return w.ServiceAccount, true, nil
			}
			// The pod controller sets it and then writes through the defaulting webhook
			// as a privileged user; that write must not lose it.
			if w.TrustedWriter {
				v, ok := w.Labels[metadata.KumaServiceAccount]
				return v, ok, nil
			}
			return "", false, nil
		},
		StoredAsAnnotation: true,
	},
	{
		Key:     mesh_proto.ListenerZoneIngressLabel,
		Owner:   OwnerControlPlane,
		Compute: listenerLabel(mesh_proto.Dataplane_Networking_Listener_ZoneIngress),
	},
	{
		Key:     mesh_proto.ListenerZoneEgressLabel,
		Owner:   OwnerControlPlane,
		Compute: listenerLabel(mesh_proto.Dataplane_Networking_Listener_ZoneEgress),
	},
	{
		Key:   metadata.KumaWorkload,
		Owner: OwnerUser,
		Compute: func(w Write, _ ControlPlane) (string, bool, error) {
			if w.Workload != "" {
				return w.Workload, true, nil
			}
			v, ok := w.Labels[metadata.KumaWorkload]
			return v, ok, nil
		},
		StoredAsAnnotation: true,
	},
}

// AllComputedLabels lists every registered key. If changed sync with:
// https://github.com/Kong/shared-speakeasy/blob/b3ddd3ef1f31e42bfe71b96ea473493072f9742c/customtypes/kumalabels/kumalabels.go#L15
var AllComputedLabels = computedLabels()

func computedLabels() map[string]struct{} {
	keys := map[string]struct{}{}
	for _, d := range registry {
		keys[d.Key] = struct{}{}
	}
	return keys
}

// AnnotationBacked returns the keys Kubernetes stores as annotations instead of labels.
func AnnotationBacked() []string {
	var keys []string
	for _, d := range registry {
		if d.StoredAsAnnotation {
			keys = append(keys, d.Key)
		}
	}
	return keys
}

func storedAsAnnotation(key string) bool {
	for _, d := range registry {
		if d.Key == key {
			return d.StoredAsAnnotation
		}
	}
	return false
}
