package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceInsightSpec defines the observed state of services in the Mesh
type ServiceInsightSpec = map[string]interface{}

// ServiceInsight is the Schema for the Service Insights API
type ServiceInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Mesh              string `json:"mesh,omitempty"`

	Spec ServiceInsightSpec `json:"spec,omitempty"`
}

// ServiceInsightList contains a list of ServiceInsight
type ServiceInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceInsight `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceInsight{}, &ServiceInsightList{})
}
