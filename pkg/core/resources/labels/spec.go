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

// Owner declares who is authoritative for a reserved label's value.
type Owner string

const (
	// OwnerControlPlane: the control plane both writes the value and knows how
	// to compute it (e.g. a zone knows the status of something). Compute
	// force-sets these from the registry, overriding whatever the user sent.
	OwnerControlPlane Owner = "control-plane"
	// OwnerUser: the user sets the value; the control plane never computes it
	// and only validates the format / allowed values.
	OwnerUser Owner = "user"
	// OwnerSystem: the value comes from another trusted component in the system
	// that the control plane cannot compute itself (e.g. global trusts that the
	// zone set the label correctly). Left untouched by Compute.
	OwnerSystem Owner = "system"
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

// register adds a spec to the registry. It panics if a key already has a spec
// whose RequiredOn could match the same context, guaranteeing that at most one
// spec per key can match.
func register(s LabelSpec) {
	if s.Owner == OwnerControlPlane && s.Expected == nil {
		panic("resource_labels: OwnerControlPlane spec must define Expected for key " + s.Key)
	}
	for _, existing := range registry[s.Key] {
		if requiredOnOverlap(existing.RequiredOn, s.RequiredOn) {
			panic("resource_labels: overlapping RequiredOn for key " + s.Key)
		}
	}
	registry[s.Key] = append(registry[s.Key], s)
}

// requiredOnOverlap reports whether two predicates could both match some
// context. It returns false only when they are provably disjoint on one of the
// single-valued dimensions (a context carries exactly one Mode, Scope,
// Environment and ResourceType); the flag/trait/bool dimensions are treated
// conservatively as always-possibly-overlapping.
func requiredOnOverlap(a, b RequiredOn) bool {
	return !disjoint(a.Modes, b.Modes) &&
		!disjoint(a.ResourceScopes, b.ResourceScopes) &&
		!disjoint(a.Environments, b.Environments) &&
		!disjoint(a.ResourceTypes, b.ResourceTypes)
}

// disjoint reports whether two allowlists can never be satisfied by the same
// single value. An empty list means "any value", so it is never disjoint.
func disjoint[T comparable](a, b []T) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	for _, x := range a {
		if slices.Contains(b, x) {
			return false
		}
	}
	return true
}

// matchedSpec returns the first spec whose RequiredOn matches. At most one spec
// per key can match, as enforced by register.
func matchedSpec(specs []LabelSpec, ctx ValidationContext) (LabelSpec, bool) {
	for _, s := range specs {
		if s.RequiredOn.Matches(ctx) {
			return s, true
		}
	}
	return LabelSpec{}, false
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
