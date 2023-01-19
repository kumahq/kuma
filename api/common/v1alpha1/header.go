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
	Name  HeaderName        `json:"name,omitempty"`
	Value HeaderValue       `json:"value,omitempty"`
}

// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=`^[A-Za-z0-9!#$%&'*+\-.^_\x60|~]+$`
type HeaderName string

type HeaderValue string
