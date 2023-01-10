// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshTrafficPermission defines permission for traffic between data planes
// proxies.
type MeshTrafficPermission struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// From list makes a match between clients and corresponding configurations
	From []From `json:"from,omitempty"`
}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// clients.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of clients referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type Action string

// ALLOW action lets the requests pass
var ALLOW Action = "ALLOW"

// DENY action blocks the requests
var DENY Action = "DENY"

// ALLOW_WITH_SHADOW_DENY action lets the requests pass but emits logs as if
// requests are denied
var ALLOW_WITH_SHADOW_DENY Action = "ALLOW_WITH_SHADOW_DENY"

type Conf struct {
	// Action defines a behavior for the specified group of clients:
	// +kubebuilder:validation:Enum=ALLOW;DENY;ALLOW_WITH_SHADOW_DENY
	Action Action `json:"action,omitempty"`
}
