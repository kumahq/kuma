package v1alpha1

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ContainerPatch stores a list of patches to apply to init and sidecar containers.
//
// +k8s:deepcopy-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=kuma,scope=Namespaced
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

// JsonPatchBlock is one json patch operation block.
type JsonPatchBlock struct {
	// Op is a jsonpatch operation string.
	// +required
	// +kubebuilder:validation:Enum=add;remove;replace;move;copy
	Op string `json:"op"`

	// Path is a jsonpatch path string.
	// +required
	Path string `json:"path"`

	// Value must be a string representing a valid json object used
	// by replace and add operations. String has to be escaped with " to be valid a json object.
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

func ToJsonPatch(in []JsonPatchBlock) jsonpatch.Patch {
	var res []jsonpatch.Operation
	for _, o := range in {
		op := json.RawMessage(fmt.Sprintf(`%q`, o.Op))
		path := json.RawMessage(fmt.Sprintf(`%q`, o.Path))
		from := json.RawMessage(fmt.Sprintf(`%q`, o.From))
		value := json.RawMessage(o.Value)
		res = append(res, jsonpatch.Operation{
			"op":    &op,
			"path":  &path,
			"from":  &from,
			"value": &value,
		})
	}
	return res
}
