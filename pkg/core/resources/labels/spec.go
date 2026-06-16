package labels

import (
	"slices"

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

// ResourceTrait names a boolean property of a resource descriptor that
// RequiredOn can require.
type ResourceTrait string

const (
	TraitPolicy           ResourceTrait = "Policy"
	TraitProxy            ResourceTrait = "Proxy"
	TraitPluginOriginated ResourceTrait = "PluginOriginated"
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
	// ResourceTraits lists descriptor traits the resource must have (all required).
	ResourceTraits []ResourceTrait
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
	for _, trait := range r.ResourceTraits {
		if !traitHolds(trait, ctx.Descriptor) {
			return false
		}
	}
	if r.RequiresNamespace && ctx.Namespace.value == "" {
		return false
	}
	return true
}

func traitHolds(t ResourceTrait, d core_model.ResourceTypeDescriptor) bool {
	switch t {
	case TraitPolicy:
		return d.IsPolicy
	case TraitProxy:
		return d.IsProxy
	case TraitPluginOriginated:
		return d.IsPluginOriginated
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
	// Expected returns the CP value when RequiredOn matches.
	// nil means the CP has no opinion on the value (any user value is accepted).
	Expected func(ctx ValidationContext) string
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
