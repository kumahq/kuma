package v1alpha1

import (
	jsonpatch "github.com/evanphx/json-patch"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ContainerPatch stores a list of patches to apply to init and sidecar containers.
//
// +k8s:deepcopy-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type ContainerPatch struct {
	kube_meta.TypeMeta   `json:",inline"`
	kube_meta.ObjectMeta `json:"metadata,omitempty"`

	Mesh string             `json:"mesh,omitempty"`
	Spec ContainerPatchSpec `json:"spec,omitempty"`
}

// ContainerPatchSpec specifies the options available for a ContainerPatch
// +k8s:deepcopy-gen=true
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

// ContainerPatchList contains a list of ContainerPatch instances
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
type ContainerPatchList struct {
	kube_meta.TypeMeta `json:",inline"`
	kube_meta.ListMeta `json:"metadata,omitempty"`
	Items              []ContainerPatch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContainerPatch{}, &ContainerPatchList{})
}

func JsonPatchBlockToPatch(patchBlock JsonPatchBlock) (jsonpatch.Patch, error) {
	patchStr := `[{"op": "` + patchBlock.Op + `", "path": "` + patchBlock.Path + `" `
	if patchBlock.Value != "" {
		// Value needs to be actual json string.
		patchStr = patchStr + `, "value": ` + patchBlock.Value
	}
	if patchBlock.From != "" {
		// Value needs to be actual json string.
		patchStr = patchStr + `, "from": "` + patchBlock.Value + `" `
	}
	patchStr += `}]`

	return jsonpatch.DecodePatch([]byte(patchStr))
}
