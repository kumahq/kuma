// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshTCPRoute
// +kuma:policy:singular_display_name=Mesh TCP Route
//
// This policy defines its own `GetDefault` method so that it can have the given
// structure for deserialization but still use the generic policy merging
// machinery.
//
// +kuma:policy:skip_get_default=true
type MeshTCPRoute struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in-place.
	TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`
	// To list makes a match between the consumed services and corresponding
	// configurations
	// +kubebuilder:validation:MinItems=1
	To []To `json:"to,omitempty"`
}

// At this point there is no plan to introduce address matching
// capabilities for `MeshTCPRoute` in foreseeable future. We try to be
// as close with structures of our policies to the Gateway API
// as possible. It means, that even if Gateway API currently doesn't
// have plans to support this kind of matching as well (ref.
// Kubernetes Gateway API GEP-735: TCP and UDP addresses matching -
// https://gateway-api.sigs.k8s.io/geps/gep-735/), its structures
// are ready to potentially support it.
//
// As a result every element of the route destination section of
// the policy configuration (`spec.to[]`) contains a `rules` property.
// This property is a list of elements, which potentially will allow
// to specify `match` configuration.
//
// Without specifying `match`es, it would be nonsensical to accept more
// than 1 `rule`.

type To struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Rules contains the routing rules applies to a combination of top-level
	// targetRef and the targetRef in this entry.
	// +kubebuilder:validation:MaxItems=1
	Rules []Rule `json:"rules,omitempty"`
}

type Rule struct {
	// Default holds routing rules that can be merged with rules from other
	// policies.
	Default RuleConf `json:"default"`
}

type RuleConf struct {
	// +kubebuilder:validation:MinItems=1
	BackendRefs []common_api.BackendRef `json:"backendRefs"`
}
