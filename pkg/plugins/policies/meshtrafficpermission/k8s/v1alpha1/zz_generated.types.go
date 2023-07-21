// Generated by tools/policy-gen
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	policy "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=kuma,scope=Namespaced
// +kubebuilder:printcolumn:name="TargetRef Kind",type="string",JSONPath=".spec.targetRef.kind"
// +kubebuilder:printcolumn:name="TargetRef Name",type="string",JSONPath=".spec.targetRef.name"
type MeshTrafficPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification of the Kuma MeshTrafficPermission resource.
	// +kubebuilder:validation:Optional
	Spec *policy.MeshTrafficPermission `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type MeshTrafficPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeshTrafficPermission `json:"items"`
}

func (cb *MeshTrafficPermission) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *MeshTrafficPermission) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *MeshTrafficPermission) GetMesh() string {
	if mesh, ok := cb.ObjectMeta.Labels[metadata.KumaMeshLabel]; ok {
		return mesh
	} else {
		return core_model.DefaultMesh
	}
}

func (cb *MeshTrafficPermission) SetMesh(mesh string) {
	if cb.ObjectMeta.Labels == nil {
		cb.ObjectMeta.Labels = map[string]string{}
	}
	cb.ObjectMeta.Labels[metadata.KumaMeshLabel] = mesh
}

func (cb *MeshTrafficPermission) GetSpec() (core_model.ResourceSpec, error) {
	return cb.Spec, nil
}

func (cb *MeshTrafficPermission) SetSpec(spec core_model.ResourceSpec) {
	if spec == nil {
		cb.Spec = nil
		return
	}

	if _, ok := spec.(*policy.MeshTrafficPermission); !ok {
		panic(fmt.Sprintf("unexpected protobuf message type %T", spec))
	}

	cb.Spec = spec.(*policy.MeshTrafficPermission)
}

func (cb *MeshTrafficPermission) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *MeshTrafficPermissionList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	SchemeBuilder.Register(&MeshTrafficPermission{}, &MeshTrafficPermissionList{})
	registry.RegisterObjectType(&policy.MeshTrafficPermission{}, &MeshTrafficPermission{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshTrafficPermission",
		},
	})
	registry.RegisterListType(&policy.MeshTrafficPermission{}, &MeshTrafficPermissionList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshTrafficPermissionList",
		},
	})
}
