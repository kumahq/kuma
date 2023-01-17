// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshHTTPRoute
// +kuma:policy:skip_registration=true
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

// +kubebuilder:validation:Enum=RequestHeaderModifier;ResponseHeaderModifier
type FilterType string

const (
	RequestHeaderModifierType  FilterType = "RequestHeaderModifier"
	ResponseHeaderModifierType FilterType = "ResponseHeaderModifier"
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

type Filter struct {
	Type                   FilterType      `json:"type"`
	RequestHeaderModifier  *HeaderModifier `json:"requestHeaderModifier,omitempty"`
	ResponseHeaderModifier *HeaderModifier `json:"responseHeaderModifier,omitempty"`
}

type BackendRef struct {
	common_api.TargetRef `json:",omitempty"`
	// +kubebuilder:validation:Minimum=0
	Weight uint `json:"weight,omitempty"`
}
