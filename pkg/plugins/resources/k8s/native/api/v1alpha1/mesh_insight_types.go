package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

// MeshInsight is the Schema for the MeshInsights API.
//
// +kubebuilder:object:root=true
type MeshInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Mesh              string `json:"mesh,omitempty"`

	Spec model.RawMessage `json:"spec,omitempty"`
}

// MeshInsightList contains a list of MeshInsights.
//
// +kubebuilder:object:root=true
type MeshInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeshInsight `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MeshInsight{}, &MeshInsightList{})
}
