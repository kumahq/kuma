// +kubebuilder:object:generate=true
package v1alpha1

// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=`^[a-z0-9!#$%&'*+\-.^_\x60|~]+$`
type HeaderName string

type HeaderValue string

type HeaderMatchType string

// HeaderMatchType constants.
const (
	HeaderMatchExact             HeaderMatchType = "Exact"
	HeaderMatchPresent           HeaderMatchType = "Present"
	HeaderMatchRegularExpression HeaderMatchType = "RegularExpression"
	HeaderMatchAbsent            HeaderMatchType = "Absent"
	HeaderMatchPrefix            HeaderMatchType = "Prefix"
)

// HeaderMatch describes how to select an HTTP route by matching HTTP request
// headers.
type HeaderMatch struct {
	// Type specifies how to match against the value of the header.
	// +optional
	// +kubebuilder:default=Exact
	// +kubebuilder:validation:Enum=Exact;Present;RegularExpression;Absent;Prefix
	// +kuma:nolint // https://github.com/kumahq/kuma/issues/14107
	Type *HeaderMatchType `json:"type,omitempty"`

	// Name is the name of the HTTP Header to be matched. Name MUST be lower case
	// as they will be handled with case insensitivity (See https://tools.ietf.org/html/rfc7230#section-3.2).
	Name HeaderName `json:"name"`

	// Value is the value of HTTP Header to be matched.
	// +kuma:nolint // https://github.com/kumahq/kuma/issues/14107
	Value HeaderValue `json:"value,omitempty"`
}
