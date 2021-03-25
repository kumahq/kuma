package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MeshInsightSpec defines the observed state of Mesh
type MeshInsightSpec = map[string]interface{}

// MeshInsight is the Schema for the Mesg Insights API
type MeshInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Mesh              string `json:"mesh,omitempty"`

	Spec MeshInsightSpec `json:"spec,omitempty"`
}

// MeshInsightList contains a list of MeshInsights
type MeshInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeshInsight `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MeshInsight{}, &MeshInsightList{})
}
