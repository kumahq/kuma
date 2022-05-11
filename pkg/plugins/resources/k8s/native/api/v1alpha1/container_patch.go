package v1alpha1

import (
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ContainerPatch represents a managed instance of a dataplane proxy for a Kuma
// Gateway.
type ContainerPatch struct {
	kube_meta.TypeMeta   `json:",inline"`
	kube_meta.ObjectMeta `json:"metadata,omitempty"`

	Spec ContainerPatchSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen=true

// ContainerPatchSpec specifies the options available for a GatewayDataplane.
type ContainerPatchSpec struct {
	// SidecarPatch specifies jsonpatch to apply to a sidecar container.
	SidecarPatch []JsonPatchBlock `json:"sidecarPatch,omitempty"`

	// InitPatch specifies jsonpatch to apply to an init container.
	InitPatch []JsonPatchBlock `json:"initPatch,omitempty"`
}

// JsonPatchBLock is one json patch operation block.
type JsonPatchBlock struct {
	// Op is a jsonpatch operation string.
	// +required
	Op string `json:"op"`

	// Path is a jsonpatch path string.
	// +required
	Path string `json:"path"`

	// Value must be a string representing a valid json object used
	// by replace and add operations.
	Value string `json:"value,omitempty"`

	// From is a jsonpatch from string, used by move and copy operations.
	From string `json:"from,omitempty"`
}

// +k8s:deepcopy-gen=true
// ContainerPatchList contains a list of GatewayInstances.
//
// +kubebuilder:object:root=true
type ContainerPatchList struct {
	kube_meta.TypeMeta `json:",inline"`
	kube_meta.ListMeta `json:"metadata,omitempty"`
	Items              []ContainerPatch `json:"items"`
}
