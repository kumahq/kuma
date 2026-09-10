package v1alpha1

import (
	"encoding/json"
	"errors"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
)

func mustMarshalJSON(spec any) []byte {
	bytes, err := json.Marshal(spec)
	if err != nil {
		panic(err)
	}
	return bytes
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=kuma,scope=Cluster,shortName=m
type Mesh struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Mesh is the name of the Kuma mesh this resource belongs to.
	// It may be omitted for cluster-scoped resources.
	//
	// +kubebuilder:validation:Optional
	Mesh string `json:"mesh,omitempty"`
	// Spec is the specification of the Kuma Mesh resource.
	// +kubebuilder:validation:Optional
	Spec *apiextensionsv1.JSON `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
type MeshList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mesh `json:"items"`
}

func init() {
	knownTypes = append(knownTypes, &Mesh{}, &MeshList{})
}

func (cb *Mesh) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *Mesh) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *Mesh) GetMesh() string {
	return cb.Mesh
}

func (cb *Mesh) SetMesh(mesh string) {
	cb.Mesh = mesh
}

func (cb *Mesh) GetSpec() (core_model.ResourceSpec, error) {
	spec := cb.Spec
	m := mesh_proto.Mesh{}

	if spec == nil || len(spec.Raw) == 0 {
		return &m, nil
	}

	err := json.Unmarshal(spec.Raw, &m)
	return &m, err
}

func (cb *Mesh) SetSpec(spec core_model.ResourceSpec) {
	if spec == nil {
		cb.Spec = nil
		return
	}

	s, ok := spec.(*mesh_proto.Mesh)
	if !ok {
		panic(fmt.Sprintf("unexpected spec type %T", spec))
	}

	cb.Spec = &apiextensionsv1.JSON{Raw: mustMarshalJSON(s)}
}

func (cb *Mesh) GetStatus() (core_model.ResourceStatus, error) {
	return nil, nil
}

func (cb *Mesh) SetStatus(_ core_model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (cb *Mesh) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *MeshList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.Mesh{}, &Mesh{
		APIVersion: GroupVersion.String(),
		Kind:       "Mesh",
	})
	registry.RegisterListType(&mesh_proto.Mesh{}, &MeshList{
		APIVersion: GroupVersion.String(),
		Kind:       "MeshList",
	})
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=kuma,scope=Cluster
type MeshInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Mesh is the name of the Kuma mesh this resource belongs to.
	// It may be omitted for cluster-scoped resources.
	//
	// +kubebuilder:validation:Optional
	Mesh string `json:"mesh,omitempty"`
	// Spec is the specification of the Kuma MeshInsight resource.
	// +kubebuilder:validation:Optional
	Spec *apiextensionsv1.JSON `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
type MeshInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeshInsight `json:"items"`
}

func init() {
	knownTypes = append(knownTypes, &MeshInsight{}, &MeshInsightList{})
}

func (cb *MeshInsight) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *MeshInsight) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *MeshInsight) GetMesh() string {
	return cb.Mesh
}

func (cb *MeshInsight) SetMesh(mesh string) {
	cb.Mesh = mesh
}

func (cb *MeshInsight) GetSpec() (core_model.ResourceSpec, error) {
	spec := cb.Spec
	m := mesh_proto.MeshInsight{}

	if spec == nil || len(spec.Raw) == 0 {
		return &m, nil
	}

	err := json.Unmarshal(spec.Raw, &m)
	return &m, err
}

func (cb *MeshInsight) SetSpec(spec core_model.ResourceSpec) {
	if spec == nil {
		cb.Spec = nil
		return
	}

	s, ok := spec.(*mesh_proto.MeshInsight)
	if !ok {
		panic(fmt.Sprintf("unexpected spec type %T", spec))
	}

	cb.Spec = &apiextensionsv1.JSON{Raw: mustMarshalJSON(s)}
}

func (cb *MeshInsight) GetStatus() (core_model.ResourceStatus, error) {
	return nil, nil
}

func (cb *MeshInsight) SetStatus(_ core_model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (cb *MeshInsight) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *MeshInsightList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.MeshInsight{}, &MeshInsight{
		APIVersion: GroupVersion.String(),
		Kind:       "MeshInsight",
	})
	registry.RegisterListType(&mesh_proto.MeshInsight{}, &MeshInsightList{
		APIVersion: GroupVersion.String(),
		Kind:       "MeshInsightList",
	})
}
