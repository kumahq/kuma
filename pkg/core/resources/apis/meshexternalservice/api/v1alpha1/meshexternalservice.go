// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshExternalService
// +kuma:policy:is_policy=false
// +kuma:policy:has_status=true
type MeshExternalService struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in-place.
	TargetRef common_api.TargetRef `json:"targetRef"`
}

type MeshExternalServiceStatus struct{}
