package labels

import (
	"slices"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// ValidationContext carries everything the registry needs to decide whether a
// reserved label applies and what its computed value should be.
type ValidationContext struct {
	Mode         config_core.CpMode
	Env          config_core.EnvironmentType
	ZoneName     string
	Namespace    Namespace
	Descriptor   core_model.ResourceTypeDescriptor
	Spec         core_model.ResourceSpec
	ResourceName string
	ResourceMesh string
}

// Owner controls a reserved label's value.
type Owner string

const (
	OwnerControlPlane Owner = "control-plane"
	OwnerUser         Owner = "user"
	OwnerSystem       Owner = "system"
)

type SpecTrait string

const (
	HasZoneIngressListener SpecTrait = "HasZoneIngressListener"
	HasZoneEgressListener  SpecTrait = "HasZoneEgressListener"
)

// RequiredOn declares when a label applies. Fields are AND-combined.
type RequiredOn struct {
	Modes             []config_core.CpMode
	ResourceScopes    []core_model.ResourceScope
	Environments      []config_core.EnvironmentType
	KDSFlags          []core_model.KDSFlagType
	ResourceTypes     []core_model.ResourceType
	SpecTraits        []SpecTrait
	Policy            bool
	RequiresNamespace bool
}

func (r RequiredOn) Matches(ctx ValidationContext) bool {
	if len(r.Modes) > 0 && !slices.Contains(r.Modes, ctx.Mode) {
		return false
	}
	if len(r.ResourceScopes) > 0 && !slices.Contains(r.ResourceScopes, ctx.Descriptor.Scope) {
		return false
	}
	if len(r.Environments) > 0 && !slices.Contains(r.Environments, ctx.Env) {
		return false
	}
	for _, flag := range r.KDSFlags {
		if !ctx.Descriptor.KDSFlags.Has(flag) {
			return false
		}
	}
	if len(r.ResourceTypes) > 0 && !slices.Contains(r.ResourceTypes, ctx.Descriptor.Name) {
		return false
	}
	if r.Policy && (!ctx.Descriptor.IsPolicy || !ctx.Descriptor.IsPluginOriginated) {
		return false
	}
	for _, trait := range r.SpecTraits {
		if !specTraitHolds(trait, ctx.Spec) {
			return false
		}
	}
	if r.RequiresNamespace && ctx.Namespace.value == "" {
		return false
	}
	return true
}

func specTraitHolds(t SpecTrait, spec core_model.ResourceSpec) bool {
	dp, ok := spec.(*mesh_proto.Dataplane)
	if !ok {
		return false
	}
	var want mesh_proto.Dataplane_Networking_Listener_Type
	switch t {
	case HasZoneIngressListener:
		want = mesh_proto.Dataplane_Networking_Listener_ZoneIngress
	case HasZoneEgressListener:
		want = mesh_proto.Dataplane_Networking_Listener_ZoneEgress
	default:
		return false
	}
	for _, l := range dp.GetNetworking().GetListeners() {
		if l.Type == want {
			return true
		}
	}
	return false
}

// LabelSpec describes one reserved label. kuma.io/origin is separate.
type LabelSpec struct {
	Key           string
	Owner         Owner
	AllowedValues []string
	RequiredOn    RequiredOn
	// nil means the CP has no opinion on the value.
	Expected func(ctx ValidationContext) (string, error)
}

var registry = map[string][]LabelSpec{}

func register(s LabelSpec) {
	if s.Owner == OwnerControlPlane && s.Expected == nil {
		panic("resource_labels: OwnerControlPlane spec must define Expected for key " + s.Key)
	}
	registry[s.Key] = append(registry[s.Key], s)
}

// matchedSpec panics on overlapping RequiredOn entries.
func matchedSpec(specs []LabelSpec, ctx ValidationContext) (LabelSpec, bool) {
	var hit LabelSpec
	found := false
	for _, s := range specs {
		if !s.RequiredOn.Matches(ctx) {
			continue
		}
		if found {
			panic("resource_labels: overlapping RequiredOn for key " + s.Key)
		}
		hit = s
		found = true
	}
	return hit, found
}

// isControlPlaneKey reports whether a registry key is control-plane-owned, and
// therefore safe for Compute to force-set or delete.
func isControlPlaneKey(specs []LabelSpec) bool {
	for _, s := range specs {
		if s.Owner == OwnerControlPlane {
			return true
		}
	}
	return false
}
