package v1alpha1

type Match struct {
	// SpiffeId defines a matcher configuration for SpiffeId matching
	SpiffeId *SpiffeIdMatch `json:"spiffeId,omitempty"`
}

// +kubebuilder:validation:Enum=Exact;PathPrefix
type SpiffeIdMatchType string

type SpiffeIdMatch struct {
	// Type defines how to match incoming traffic by SpiffeId. `Exact` or `PathPrefix` are allowed.
	Type SpiffeIdMatchType `json:"type"`
	// Value is SpiffeId of a client that needs to match for the configuration to be applied
	Value string `json:"value"`
}
