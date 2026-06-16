package labels

import (
	"slices"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

// Owner classifies who controls a reserved label's value.
type Owner string

const (
	// OwnerControlPlane is computed by the CP.
	OwnerControlPlane Owner = "control-plane"
	// OwnerUser is set by users, optionally with constrained values.
	OwnerUser Owner = "user"
	// OwnerSystem is set only by trusted CP-internal flows.
	OwnerSystem Owner = "system"
)

// SpecTrait names a boolean property of a resource's spec that RequiredOn can
// require.
type SpecTrait string

const (
	HasZoneIngressListener SpecTrait = "HasZoneIngressListener"
	HasZoneEgressListener  SpecTrait = "HasZoneEgressListener"
)

// RequiredOn declares the contexts where a label applies.
// All fields are AND-combined. A zero-valued field places no constraint.
type RequiredOn struct {
	// Modes is a CP-mode allowlist; ctx.Mode must equal one entry.
	Modes []config_core.CpMode
	// ResourceScopes is a resource-scope allowlist.
	ResourceScopes []core_model.ResourceScope
	// Environments is a CP-environment allowlist (Kubernetes/Universal).
	Environments []config_core.EnvironmentType
	// KDSFlags lists KDS flags the descriptor must carry (all required).
	KDSFlags []core_model.KDSFlagType
	// ResourceTypes is a resource-type allowlist; ctx.Descriptor.Name must equal one entry.
	ResourceTypes []core_model.ResourceType
	// SpecTraits lists spec properties the resource must have (all required).
	SpecTraits []SpecTrait
	// Policy requires the resource to be a plugin-originated policy.
	Policy bool
	// RequiresNamespace requires the resource to be namespace-scoped.
	RequiresNamespace bool
}

// Matches reports whether ctx satisfies every constraint expressed by r.
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

// LabelSpec describes one reserved label. kuma.io/origin is handled separately.
type LabelSpec struct {
	Key           string
	Owner         Owner
	AllowedValues []string
	// RequiredOn declares when the label applies. Zero value = always applies.
	RequiredOn RequiredOn
	// Expected returns the CP value when RequiredOn matches. An error means the
	// resource is malformed and no canonical value can be computed.
	// nil means the CP has no opinion on the value (any user value is accepted).
	Expected func(ctx ValidationContext) (string, error)
}

var registry = map[string][]LabelSpec{}

func register(s LabelSpec) {
	if s.Owner == OwnerControlPlane && s.Expected == nil {
		panic("resource_labels: OwnerControlPlane spec must define Expected for key " + s.Key)
	}
	registry[s.Key] = append(registry[s.Key], s)
}

// matchedSpec returns the unique spec from specs whose RequiredOn matches ctx.
// Panics if two specs both match: that means the registry has overlapping
// RequiredOn for the same key, which is a programmer error.
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
