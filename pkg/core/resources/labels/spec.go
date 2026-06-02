package labels

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
)

// Owner classifies who controls a reserved label's value.
type Owner string

const (
	// OwnerControlPlane: the CP computes the value. Users may set the label only
	// if their value matches the CP-computed value.
	OwnerControlPlane Owner = "control-plane"
	// OwnerUser: free-form user flag. Optionally constrained to a set of values.
	OwnerUser Owner = "user"
	// OwnerSystem: managed by CP-internal flows only (KDS sync, lifecycle).
	// User input is always rejected.
	OwnerSystem Owner = "system"
)

// LabelSpec is the declarative entry for one reserved label.
//
// Descriptive fields (Key, Description, Owner, AllowedValues, ExpectedValueExpr,
// AppliesToExpr, RequiredWhenExpr) are JSON-serializable and form the catalog
// returned by Schema(); a future introspection endpoint can expose them as-is.
//
// Behavioral fields (Expected, RequirePresence) are not serialized; they are
// the runtime hooks consulted by Validate.
type LabelSpec struct {
	Key               string   `json:"key"`
	Description       string   `json:"description,omitempty"`
	Owner             Owner    `json:"owner"`
	AllowedValues     []string `json:"allowedValues,omitempty"`
	ExpectedValueExpr string   `json:"expectedValueExpr,omitempty"`
	AppliesToExpr     string   `json:"appliesTo,omitempty"`
	RequiredWhenExpr  string   `json:"requiredWhen,omitempty"`
	// OpenValue: when Owner == OwnerControlPlane and Expected returns
	// applies=true, the user-supplied value is accepted without an equality
	// check. Used for CP-managed labels whose value is opaque to validation
	// (e.g. kuma.io/workload on Universal Dataplanes — value comes from the
	// user but the label only applies to specific resource types).
	OpenValue bool `json:"openValue,omitempty"`

	// Expected returns the value the CP would compute for ctx. applies=false
	// means the label is not applicable in this context: any user-provided
	// value is rejected. Only used when Owner == OwnerControlPlane.
	Expected func(ctx ValidationContext) (value string, applies bool) `json:"-"`

	// RequirePresence returns true iff the label MUST be set in ctx.
	// nil means never required. Only used when Owner == OwnerControlPlane.
	RequirePresence func(ctx ValidationContext) bool `json:"-"`
}

var registry = map[string]LabelSpec{}

// register adds a LabelSpec to the registry. Called from per-label init()s.
// Panics on duplicate Key so init-order bugs are surfaced loudly during tests.
func register(s LabelSpec) {
	if _, dup := registry[s.Key]; dup {
		panic("resource_labels: duplicate registration for " + s.Key)
	}
	registry[s.Key] = s
}

// Schema returns the registered label specs as a JSON-serializable slice.
// The slice is intentionally unordered — callers that care about ordering
// should sort by Key.
func Schema() []LabelSpec {
	out := make([]LabelSpec, 0, len(registry))
	for _, s := range registry {
		out = append(out, s)
	}
	return out
}

// IsRegistered reports whether key has a LabelSpec in the registry.
// Useful for tests and for the "unknown reserved key" path in Validate.
func IsRegistered(key string) bool {
	_, ok := registry[key]
	return ok
}

// IsReservedLabelKey is re-exported here to keep callers free from the
// api/mesh/v1alpha1 import path.
func IsReservedLabelKey(key string) bool {
	return mesh_proto.IsReservedLabelKey(key)
}
