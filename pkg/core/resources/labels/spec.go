package labels

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
type LabelSpec struct {
	Key           string
	Owner         Owner
	AllowedValues []string
	// OpenValue: when Owner == OwnerControlPlane and Expected returns
	// applies=true, the user-supplied value is accepted without an equality
	// check. Used for CP-managed labels whose value is opaque to validation
	// (e.g. kuma.io/workload on Universal Dataplanes — value comes from the
	// user but the label only applies to specific resource types).
	OpenValue bool
	// AllowAnyWhenNotApplicable: when Owner == OwnerControlPlane and Expected
	// returns applies=false, accept any AllowedValues value the user supplies
	// instead of rejecting it as "reserved". Used by kuma.io/origin: contexts
	// where the CP doesn't strictly enforce a specific origin (e.g. user
	// namespaces on K8s zones) still want the vocabulary checked.
	AllowAnyWhenNotApplicable bool
	// StrictMatch: when Owner == OwnerControlPlane, Expected returns
	// applies=true, and the user-supplied value mismatches expected, surface
	// an error instead of a warning. Used by kuma.io/origin so a wrong
	// 'global'/'zone' is not silently overridden — that would mask the real
	// failure mode (resource being applied to the wrong CP). Other CP-owned
	// labels stay non-strict: Compute regenerates them and a warning is
	// enough.
	StrictMatch bool

	// Expected returns the value the CP would compute for ctx. applies=false
	// means the label is not applicable in this context: any user-provided
	// value is rejected. Only used when Owner == OwnerControlPlane.
	Expected func(ctx ValidationContext) (value string, applies bool)

	// RequirePresence returns true iff the label MUST be set in ctx.
	// nil means never required. Only used when Owner == OwnerControlPlane.
	RequirePresence func(ctx ValidationContext) bool
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
