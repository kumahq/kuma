package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

// ServiceInsight is the Schema for the ServiceInsight API.
//
// +kubebuilder:object:root=true
type ServiceInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Mesh              string `json:"mesh,omitempty"`

	Spec model.RawMessage `json:"spec,omitempty"`
}

// ServiceInsightList contains a list of ServiceInsights.
//
// +kubebuilder:object:root=true
type ServiceInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceInsight `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceInsight{}, &ServiceInsightList{})
}
