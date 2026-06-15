package labels

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

// LabelSpec describes one reserved label. kuma.io/origin is handled separately.
type LabelSpec struct {
	Key           string
	Owner         Owner
	AllowedValues []string
	// OpenValue accepts any value when Expected says the label applies.
	OpenValue bool

	// Expected returns the CP value, or applies=false when the label does not apply.
	Expected func(ctx ValidationContext) (value string, applies bool)
}

var registry = map[string]LabelSpec{}

func register(s LabelSpec) {
	if _, dup := registry[s.Key]; dup {
		panic("resource_labels: duplicate registration for " + s.Key)
	}
	registry[s.Key] = s
}
