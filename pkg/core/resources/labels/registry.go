package labels

import (
	"fmt"
	"slices"
	"strings"

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
	Compute func(key string, w Write, cp ControlPlane) (value string, ok bool, err error)

	// EnforceOnRead recomputes the label on every read. ok=false means "produce
	// nothing", never "remove".
	EnforceOnRead func(key string, r StoredResource, cp ControlPlane) (value string, ok bool)

	// ValidateFormat checks the value's syntax for any writer.
	ValidateFormat func(key, value string) []string

	// StoredAsAnnotation marks labels whose values carry a resource name and therefore
	// can be up to 253 characters, which does not fit the 63-character Kubernetes label
	// value limit. They are stored as annotations on Kubernetes and validated as names.
	StoredAsAnnotation bool
}

func keep(w Write, key string) (string, bool, error) {
	v, ok := w.Labels[key]
	return v, ok, nil
}

func keepForTrustedWriter(key string, w Write, _ ControlPlane) (string, bool, error) {
	if !w.TrustedWriter {
		return "", false, nil
	}
	return keep(w, key)
}

func oneOf(values ...string) func(string, string) []string {
	return func(key, v string) []string {
		if slices.Contains(values, v) {
			return nil
		}
		return []string{fmt.Sprintf("label %q must be %s, got %q", key, strings.Join(values, " or "), v)}
	}
}

func listenerLabel(listenerType mesh_proto.Dataplane_Networking_Listener_Type) func(string, Write, ControlPlane) (string, bool, error) {
	return func(_ string, w Write, _ ControlPlane) (string, bool, error) {
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

func computeZone(_ string, w Write, cp ControlPlane) (string, bool, error) {
	if cp.Mode == config_core.Zone && w.Descriptor.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
		return cp.Zone, cp.Zone != "", nil
	}
	return "", false, nil
}

func enforceZone(_ string, r StoredResource, cp ControlPlane) (string, bool) {
	if cp.Mode == config_core.Zone && r.IsLocal && cp.Zone != "" && r.Descriptor.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
		return cp.Zone, true
	}
	return "", false
}

// zoneOfWrite and zoneOfStored give the kuma.io/zone the resource ends up with, for
// the rules that need to compare it with a value in the spec. The zone descriptor
// runs before them in the registry but its result is not threaded through, so they
// ask it again rather than read a label the write may not carry yet.
func zoneOfWrite(w Write, cp ControlPlane) string {
	v, ok, err := computeZone(mesh_proto.ZoneTag, w, cp)
	if err != nil || !ok {
		return ""
	}
	return v
}

func zoneOfStored(r StoredResource, cp ControlPlane) string {
	if v, ok := enforceZone(mesh_proto.ZoneTag, r, cp); ok {
		return v
	}
	return r.Labels[mesh_proto.ZoneTag]
}

// Iteration order is the order the API server reports ownership violations in.
var registry = []Descriptor{
	{
		Key:   mesh_proto.ResourceOriginLabel,
		Owner: OwnerControlPlane,
		Compute: func(_ string, _ Write, cp ControlPlane) (string, bool, error) {
			switch cp.Mode {
			case config_core.Global:
				return string(mesh_proto.GlobalResourceOrigin), true, nil
			case config_core.Zone:
				return string(mesh_proto.ZoneResourceOrigin), true, nil
			default:
				return "", false, nil
			}
		},
		EnforceOnRead: func(_ string, r StoredResource, cp ControlPlane) (string, bool) {
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
		Key:           mesh_proto.ZoneTag,
		Owner:         OwnerControlPlane,
		Compute:       computeZone,
		EnforceOnRead: enforceZone,
	},
	{
		Key:   metadata.KumaMeshLabel,
		Owner: OwnerControlPlane,
		Compute: func(_ string, w Write, _ ControlPlane) (string, bool, error) {
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
		Compute: func(_ string, w Write, cp ControlPlane) (string, bool, error) {
			policy, ok := w.Spec.(core_model.Policy)
			if !w.Descriptor.IsPolicy || !w.Descriptor.IsPluginOriginated || !ok {
				return "", false, nil
			}
			role, err := ComputePolicyRole(policy, w.Namespace, zoneOfWrite(w, cp))
			if err != nil {
				return "", false, err
			}
			return string(role), true, nil
		},
		EnforceOnRead: func(_ string, r StoredResource, cp ControlPlane) (string, bool) {
			if r.Namespace.value == "" || r.Namespace.system || !r.Descriptor.IsPolicy || !r.Descriptor.IsPluginOriginated {
				return "", false
			}
			policy, ok := r.Spec.(core_model.Policy)
			if !ok {
				return "", false
			}
			role, err := ComputePolicyRole(policy, r.Namespace, zoneOfStored(r, cp))
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
		Compute: func(_ string, w Write, _ ControlPlane) (string, bool, error) {
			return w.DisplayName, true, nil
		},
		StoredAsAnnotation: true,
	},
	{
		Key:   mesh_proto.EnvTag,
		Owner: OwnerControlPlane,
		Compute: func(_ string, w Write, cp ControlPlane) (string, bool, error) {
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
		Compute: func(_ string, w Write, cp ControlPlane) (string, bool, error) {
			if !cp.IsK8s || w.Namespace.value == "" {
				return "", false, nil
			}
			return w.Namespace.value, true, nil
		},
		EnforceOnRead: func(_ string, r StoredResource, _ ControlPlane) (string, bool) {
			if r.Namespace.value != "" && !r.Namespace.system {
				return r.Namespace.value, true
			}
			return "", false
		},
	},
	{
		Key:   metadata.KumaServiceAccount,
		Owner: OwnerControlPlane,
		Compute: func(key string, w Write, cp ControlPlane) (string, bool, error) {
			if !cp.IsK8s {
				return "", false, nil
			}
			if w.ServiceAccount != "" {
				return w.ServiceAccount, true, nil
			}
			// The pod controller sets it and then writes through the defaulting webhook
			// as a privileged user; that write must not lose it.
			return keepForTrustedWriter(key, w, cp)
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
		Compute: func(_ string, w Write, _ ControlPlane) (string, bool, error) {
			if w.Workload != "" {
				return w.Workload, true, nil
			}
			return keep(w, metadata.KumaWorkload)
		},
		StoredAsAnnotation: true,
	},
	{
		Key:     mesh_proto.ManagedByLabel,
		Owner:   OwnerControlPlane,
		Compute: keepForTrustedWriter,
	},
	{
		Key:     mesh_proto.DeletionGracePeriodStartedLabel,
		Owner:   OwnerControlPlane,
		Compute: keepForTrustedWriter,
	},
	{
		Key:     metadata.KumaServiceName,
		Owner:   OwnerControlPlane,
		Compute: keepForTrustedWriter,
	},
	{
		Key:     metadata.HeadlessService,
		Owner:   OwnerControlPlane,
		Compute: keepForTrustedWriter,
	},
	{
		Key:            mesh_proto.KDSSyncLabel,
		Owner:          OwnerUser,
		ValidateFormat: oneOf("enabled", "disabled"),
	},
	{
		Key:            mesh_proto.EffectLabel,
		Owner:          OwnerUser,
		ValidateFormat: oneOf("shadow"),
	},
}

// AllComputedLabels lists every key the control plane writes itself. If changed sync with:
// https://github.com/Kong/shared-speakeasy/blob/b3ddd3ef1f31e42bfe71b96ea473493072f9742c/customtypes/kumalabels/kumalabels.go#L15
var AllComputedLabels = computedLabels()

func computedLabels() map[string]struct{} {
	keys := map[string]struct{}{}
	for _, d := range registry {
		if d.Compute != nil {
			keys[d.Key] = struct{}{}
		}
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

func lookup(key string) (Descriptor, bool) {
	for _, d := range registry {
		if d.Key == key {
			return d, true
		}
	}
	return Descriptor{}, false
}

func storedAsAnnotation(key string) bool {
	d, _ := lookup(key)
	return d.StoredAsAnnotation
}
