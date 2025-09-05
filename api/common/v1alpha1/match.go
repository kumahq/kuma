package v1alpha1

type Match struct {
	// SpiffeID defines a matcher configuration for SpiffeID matching
	SpiffeID *SpiffeIDMatch `json:"spiffeID,omitempty"`
}

// +kubebuilder:validation:Enum=Exact;Prefix
type SpiffeIDMatchType string

const (
	ExactMatchType  SpiffeIDMatchType = "Exact"
	PrefixMatchType SpiffeIDMatchType = "Prefix"
)

type SpiffeIDMatch struct {
	// Type defines how to match incoming traffic by SpiffeID. `Exact` or `Prefix` are allowed.
	Type SpiffeIDMatchType `json:"type"`
	// Value is SpiffeId of a client that needs to match for the configuration to be applied
	Value string `json:"value"`
}
