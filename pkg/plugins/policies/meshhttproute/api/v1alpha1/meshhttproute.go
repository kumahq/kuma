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
	Matches []Match `json:"matches"`
	// Default holds routing rules that can be merged with rules from other
	// policies.
	Default RuleConf `json:"default"`
}

type Match struct {
	Path PathMatch `json:"path,omitempty"`
}

type PathMatch struct {
	Prefix string `json:"prefix,omitempty"`
}

type RuleConf struct {
	Filters     *[]Filter     `json:"filters,omitempty"`
	BackendRefs *[]BackendRef `json:"backendRefs,omitempty"`
}

type Filter struct {
}

type BackendRef struct {
	common_api.TargetRef `json:",omitempty"`
	// +kubebuilder:validation:Minimum=0
	Weight uint `json:"weight,omitempty"`
}
