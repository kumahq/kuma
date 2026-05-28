package v1alpha1

type Match struct {
	// SpiffeID defines a matcher configuration for SpiffeID matching
	SpiffeID *SpiffeIDMatch `json:"spiffeID,omitempty"`
	// SNI defines a matcher configuration for matching by SNI value carried on the TLS connection
	SNI *SNIMatch `json:"sni,omitempty"`
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
	// Value is SpiffeID of a client that needs to match for the configuration to be applied
	Value string `json:"value"`
}

// +kubebuilder:validation:Enum=Exact
type SNIMatchType string

const SNIExactMatchType SNIMatchType = "Exact"

type SNIMatch struct {
	// Type defines how to match traffic by SNI. Only `Exact` is supported.
	Type SNIMatchType `json:"type"`
	// Value is the SNI carried on the TLS connection that needs to match for the configuration to be applied
	Value string `json:"value"`
}
