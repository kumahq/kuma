package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type NetworkAttachmentDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NetworkAttachmentDefinitionSpec `json:"spec"`
}

type NetworkAttachmentDefinitionSpec struct {
	Config string `json:"config"`
}

type NetworkAttachmentDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NetworkAttachmentDefinition `json:"items"`
}
