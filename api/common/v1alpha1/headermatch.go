package v1alpha1

type HeaderMatcherType string

var (
	REGULAR_EXPRESSION HeaderMatcherType = "REGULAR_EXPRESSION"
	EXACT              HeaderMatcherType = "EXACT"
	PREFIX             HeaderMatcherType = "PREFIX"
)

type HeaderMatcher struct {
	// +kubebuilder:validation:Enum=REGULAR_EXPRESSION;EXACT;PREFIX
	Type  HeaderMatcherType `json:"type,omitempty"`
	Name  string            `json:"name,omitempty"`
	Value string            `json:"value,omitempty"`
}
