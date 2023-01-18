// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshHTTPRoute
// +kuma:policy:singular_display_name=Mesh HTTP Route
//
// This policy defines its own `GetDefault` method so that it can have the given
// structure for deserialization but still use the generic policy merging
// machinery.
//
// +kuma:policy:skip_get_default=true
type MeshHTTPRoute struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef common_api.TargetRef `json:"targetRef,omitempty"`

	// To matches destination services of requests and holds configuration.
	To []To `json:"to,omitempty"`
}

type To struct {
	// TargetRef is a reference to the resource that represents a group of
	// request destinations.
	TargetRef common_api.TargetRef `json:"targetRef,omitempty"`
	// Rules contains the routing rules applies to a combination of top-level
	// targetRef and the targetRef in this entry.
	Rules []Rule `json:"rules,omitempty"`
}

type Rule struct {
	Matches []Match `json:"matches" policyMerge:"mergeKey"`
	// Default holds routing rules that can be merged with rules from other
	// policies.
	Default RuleConf `json:"default"`
}

type Match struct {
	Path   *PathMatch `json:"path,omitempty"`
	Method *Method    `json:"method,omitempty"`
}

// +kubebuilder:validation:Enum=Exact;Prefix;RegularExpression
type PathMatchType string

// +kubebuilder:validation:Enum=CONNECT;DELETE;GET;HEAD;OPTIONS;PATCH;POST;PUT;TRACE
type Method string

const (
	Exact             PathMatchType = "Exact"
	Prefix            PathMatchType = "Prefix"
	RegularExpression PathMatchType = "RegularExpression"
)

type PathMatch struct {
	// Exact or prefix matches must be an absolute path. A prefix matches only
	// if separated by a slash or the entire path.
	// +kubebuilder:validation:MinLength=1
	Value string        `json:"value"`
	Type  PathMatchType `json:"type"`
}

type RuleConf struct {
	Filters     *[]Filter     `json:"filters,omitempty"`
	BackendRefs *[]BackendRef `json:"backendRefs,omitempty"`
}

// +kubebuilder:validation:Enum=RequestHeaderModifier;ResponseHeaderModifier;RequestRedirect
type FilterType string

const (
	RequestHeaderModifierType  FilterType = "RequestHeaderModifier"
	ResponseHeaderModifierType FilterType = "ResponseHeaderModifier"
	RequestRedirectType        FilterType = "RequestRedirect"
)

// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=`^[A-Za-z0-9!#$%&'*+\-.^_\x60|~]+$`
type HeaderName string

type HeaderKeyValue struct {
	Name  HeaderName `json:"name"`
	Value string     `json:"value"`
}

// Only one action is supported per header name.
// Configuration to set or add multiple values for a header must use RFC 7230
// header value formatting, separating each value with a comma.
type HeaderModifier struct {
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Set []HeaderKeyValue `json:"set,omitempty"`
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Add []HeaderKeyValue `json:"add,omitempty"`
	// +kubebuilder:validation:MaxItems=16
	Remove []string `json:"remove,omitempty"`
}

// PreciseHostname is the fully qualified domain name of a network host. This
// matches the RFC 1123 definition of a hostname with 1 notable exception that
// numeric IP addresses are not allowed.
//
// Note that as per RFC1035 and RFC1123, a *label* must consist of lower case
// alphanumeric characters or '-', and must start and end with an alphanumeric
// character. No other punctuation is allowed.
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
type PreciseHostname string

// PortNumber defines a network port.
//
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=65535
type PortNumber int32

type RequestRedirect struct {
	// +kubebuilder:validation:Enum=http;https
	Scheme   *string          `json:"scheme,omitempty"`
	Hostname *PreciseHostname `json:"hostname,omitempty"`
	// Port is the port to be used in the value of the `Location`
	// header in the response.
	// When empty, port (if specified) of the request is used.
	//
	Port *PortNumber `json:"port,omitempty"`
	// StatusCode is the HTTP status code to be used in response.
	//
	// +kubebuilder:default=302
	// +kubebuilder:validation:Enum=301;302;303;307;308
	StatusCode *int `json:"statusCode,omitempty"`
}

type Filter struct {
	Type                   FilterType       `json:"type"`
	RequestHeaderModifier  *HeaderModifier  `json:"requestHeaderModifier,omitempty"`
	ResponseHeaderModifier *HeaderModifier  `json:"responseHeaderModifier,omitempty"`
	RequestRedirect        *RequestRedirect `json:"requestRedirect,omitempty"`
	// TODO: add path to redirect after adding URL rewrite
}

type BackendRef struct {
	common_api.TargetRef `json:",omitempty"`
	// +kubebuilder:validation:Minimum=0
	Weight uint `json:"weight,omitempty"`
}
