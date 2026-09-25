package labels

import (
	"fmt"
	"slices"
	"strings"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/version"
)

type Owner int

const (
	// OwnerControlPlane: the control plane decides the value on every write. A value
	// supplied by an untrusted writer is rejected by ValidateValue where a rule exists,
	// replaced or removed by Compute everywhere else.
	OwnerControlPlane Owner = iota
	// OwnerUser: the user sets the value; only ValidateFormat applies.
	OwnerUser
)

type Descriptor struct {
	Key   string
	Owner Owner

	// Compute returns the value to store on a write; ok=false removes the label.
	// Returning keep stores whatever was submitted. err aborts the whole write.
	Compute func(key string, w Write, cp ControlPlane) (value string, ok bool, err error)

	// EnforceOnRead recomputes the label on every read. ok=false means "produce
	// nothing", never "remove".
	EnforceOnRead func(key string, r StoredResource, cp ControlPlane) (value string, ok bool)

	// ValidateValue checks a value an untrusted writer supplied for an OwnerControlPlane
	// label.
	ValidateValue func(key, value string, w Write, cp ControlPlane) []string

	// ValidateFormat checks the value's syntax for any writer.
	ValidateFormat func(key, value string) []string

	// ValidateUpdate checks the value an untrusted update stores against the one
	// stored before it.
	ValidateUpdate func(previous, value string, w Write, cp ControlPlane) []string

	// StoredAsAnnotation marks labels whose values carry a resource name and therefore
	// can be up to 253 characters, which does not fit the 63-character Kubernetes label
	// value limit. They are stored as annotations on Kubernetes and validated as names.
	StoredAsAnnotation bool
}

func keep(w Write, key string) (string, bool, error) {
	v, ok := w.Labels[key]
	return v, ok, nil
}

func k8sWrongValueMsg(key, expected, actual string) string {
	return fmt.Sprintf("'%s' label should have '%s' value, got '%s'", key, expected, actual)
}

func keepForTrustedWriter(key string, w Write, _ ControlPlane) (string, bool, error) {
	if !w.TrustedWriter {
		return "", false, nil
	}
	return keep(w, key)
}

func rejectManualValue(key, _ string, _ Write, _ ControlPlane) []string {
	return []string{fmt.Sprintf("label %q is set by the control plane and cannot be set manually", key)}
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
	return func(key string, w Write, _ ControlPlane) (string, bool, error) {
		if !w.Descriptor.IsProxy {
			return keep(w, key)
		}
		dp, ok := w.Spec.(*mesh_proto.Dataplane)
		if !ok {
			return keep(w, key)
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
	return keep(w, mesh_proto.ZoneTag)
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
		Compute: func(_ string, w Write, cp ControlPlane) (string, bool, error) {
			switch cp.Mode {
			case config_core.Global:
				return string(mesh_proto.GlobalResourceOrigin), true, nil
			case config_core.Zone:
				return string(mesh_proto.ZoneResourceOrigin), true, nil
			default:
				return keep(w, mesh_proto.ResourceOriginLabel)
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
		ValidateValue: func(_ string, v string, w Write, cp ControlPlane) []string {
			if v == "" {
				return nil
			}
			global := string(mesh_proto.GlobalResourceOrigin)
			zone := string(mesh_proto.ZoneResourceOrigin)
			if cp.IsK8s {
				if !w.Descriptor.IsPluginOriginated || (cp.Mode != config_core.Global && !cp.FederatedZone) {
					return nil
				}
				switch {
				case cp.Mode == config_core.Global && v == zone:
					return []string{k8sWrongValueMsg(mesh_proto.ResourceOriginLabel, global, v)}
				case cp.Mode == config_core.Zone && w.Namespace.system && v != zone:
					return []string{k8sWrongValueMsg(mesh_proto.ResourceOriginLabel, zone, v)}
				}
				return nil
			}
			switch {
			case cp.Mode == config_core.Global && v != global:
				return []string{fmt.Sprintf("the origin label must be set to '%s'", global)}
			case cp.FederatedZone && v != zone:
				return []string{fmt.Sprintf("the origin label must be set to '%s'", zone)}
			}
			return nil
		},
		ValidateFormat: func(_ string, v string) []string {
			if v == "" {
				return nil
			}
			if err := mesh_proto.ResourceOrigin(v).IsValid(); err != nil {
				return []string{err.Error()}
			}
			return nil
		},
		// Without this a zone user could take over a Global-synced policy by re-applying
		// it: the write recomputes the origin as local and the Global->Zone KDS stream
		// then wedges on AlreadyExists.
		ValidateUpdate: func(previous, v string, _ Write, cp ControlPlane) []string {
			if v == previous {
				return nil
			}
			if cp.IsK8s {
				// a non-federated zone owns everything in its store
				if cp.Mode != config_core.Global && !cp.FederatedZone {
					return nil
				}
				return []string{fmt.Sprintf("'%s' label is immutable, cannot be changed from '%s' to '%s'", mesh_proto.ResourceOriginLabel, previous, v)}
			}
			return []string{fmt.Sprintf("is immutable, cannot be changed from %q to %q", previous, v)}
		},
	},
	{
		Key:           mesh_proto.ZoneTag,
		Owner:         OwnerControlPlane,
		Compute:       computeZone,
		EnforceOnRead: enforceZone,
		ValidateValue: func(_ string, v string, w Write, cp ControlPlane) []string {
			if cp.IsK8s {
				if !w.Descriptor.IsPluginOriginated || (cp.Mode != config_core.Global && !cp.FederatedZone) {
					return nil
				}
				if cp.Mode == config_core.Zone && w.Labels[mesh_proto.ResourceOriginLabel] == string(mesh_proto.ZoneResourceOrigin) && v != cp.Zone {
					return []string{k8sWrongValueMsg(mesh_proto.ZoneTag, cp.Zone, v)}
				}
				return nil
			}
			if cp.Mode == config_core.Global {
				return []string{fmt.Sprintf("%s is not allowed on a global control plane", mesh_proto.ZoneTag)}
			}
			if v != cp.Zone {
				return []string{fmt.Sprintf("%s label should have %s value", mesh_proto.ZoneTag, cp.Zone)}
			}
			return nil
		},
	},
	{
		Key:   metadata.KumaMeshLabel,
		Owner: OwnerControlPlane,
		Compute: func(_ string, w Write, _ ControlPlane) (string, bool, error) {
			if v, ok := w.Labels[metadata.KumaMeshLabel]; ok || w.Descriptor.Scope != core_model.ScopeMesh {
				return v, ok, nil
			}
			if w.Mesh != "" {
				return w.Mesh, true, nil
			}
			return core_model.DefaultMesh, true, nil
		},
		ValidateValue: func(_ string, v string, w Write, cp ControlPlane) []string {
			if cp.IsK8s || v == w.Mesh {
				return nil
			}
			return []string{fmt.Sprintf("%s label must not differ from mesh set on resource", metadata.KumaMeshLabel)}
		},
	},
	{
		Key:   mesh_proto.PolicyRoleLabel,
		Owner: OwnerControlPlane,
		Compute: func(_ string, w Write, cp ControlPlane) (string, bool, error) {
			if w.Namespace.value == "" || !w.Descriptor.IsPolicy || !w.Descriptor.IsPluginOriginated {
				return keep(w, mesh_proto.PolicyRoleLabel)
			}
			role, err := ComputePolicyRole(w.Spec.(core_model.Policy), w.Namespace, zoneOfWrite(w, cp))
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
		ValidateValue: func(_ string, v string, w Write, cp ControlPlane) []string {
			if cp.IsK8s || !w.Descriptor.IsPluginOriginated || !w.Descriptor.IsPolicy {
				return nil
			}
			if v == "" || v == string(mesh_proto.SystemPolicyRole) {
				return nil
			}
			return []string{fmt.Sprintf("%s label should have %s value, got %s", mesh_proto.PolicyRoleLabel, mesh_proto.SystemPolicyRole, v)}
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
				return keep(w, mesh_proto.EnvTag)
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
			if !cp.IsK8s {
				return "", false, nil
			}
			if w.Namespace.value != "" {
				return w.Namespace.value, true, nil
			}
			return keep(w, mesh_proto.KubeNamespaceTag)
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
		Compute: func(_ string, w Write, cp ControlPlane) (string, bool, error) {
			if !cp.IsK8s {
				return "", false, nil
			}
			if w.ServiceAccount != "" {
				return w.ServiceAccount, true, nil
			}
			return keep(w, metadata.KumaServiceAccount)
		},
		ValidateValue: func(_, _ string, w Write, cp ControlPlane) []string {
			if !cp.IsK8s || w.Descriptor.Name != core_mesh.DataplaneType {
				return nil
			}
			return []string{fmt.Sprintf("Label %q is managed by %s and cannot be set manually.", metadata.KumaServiceAccount, version.Product)}
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
		Key:           mesh_proto.ManagedByLabel,
		Owner:         OwnerControlPlane,
		Compute:       keepForTrustedWriter,
		ValidateValue: rejectManualValue,
	},
	{
		Key:           mesh_proto.DeletionGracePeriodStartedLabel,
		Owner:         OwnerControlPlane,
		Compute:       keepForTrustedWriter,
		ValidateValue: rejectManualValue,
	},
	{
		Key:           metadata.KumaServiceName,
		Owner:         OwnerControlPlane,
		Compute:       keepForTrustedWriter,
		ValidateValue: rejectManualValue,
	},
	{
		Key:           metadata.HeadlessService,
		Owner:         OwnerControlPlane,
		Compute:       keepForTrustedWriter,
		ValidateValue: rejectManualValue,
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
