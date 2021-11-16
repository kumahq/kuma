package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

// GatewayRoute is the Schema for the GatewayRoute.
//
// +kubebuilder:object:root=true
type GatewayRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Mesh string           `json:"mesh,omitempty"`
	Spec model.RawMessage `json:"spec,omitempty"`
}

// GatewayRouteList contains a list of GatewayRoutes.
//
// +kubebuilder:object:root=true
type GatewayRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GatewayRoute `json:"items"`
}
