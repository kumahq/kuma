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

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=kuma,scope=Namespaced,shortName=dp
type Dataplane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Mesh is the name of the Kuma mesh this resource belongs to.
	// It may be omitted for cluster-scoped resources.
	//
	// +kubebuilder:validation:Optional
	Mesh string `json:"mesh,omitempty"`
	// Spec is the specification of the Kuma Dataplane resource.
	// +kubebuilder:validation:Optional
	Spec *apiextensionsv1.JSON `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type DataplaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dataplane `json:"items"`
}

func init() {
	knownTypes = append(knownTypes, &Dataplane{}, &DataplaneList{})
}

func (cb *Dataplane) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *Dataplane) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *Dataplane) GetMesh() string {
	return cb.Mesh
}

func (cb *Dataplane) SetMesh(mesh string) {
	cb.Mesh = mesh
}

func (cb *Dataplane) GetSpec() (core_model.ResourceSpec, error) {
	spec := cb.Spec
	m := mesh_proto.Dataplane{}

	if spec == nil || len(spec.Raw) == 0 {
		return &m, nil
	}

	err := json.Unmarshal(spec.Raw, &m)
	return &m, err
}

func (cb *Dataplane) SetSpec(spec core_model.ResourceSpec) {
	if spec == nil {
		cb.Spec = nil
		return
	}

	s, ok := spec.(*mesh_proto.Dataplane)
	if !ok {
		panic(fmt.Sprintf("unexpected spec type %T", spec))
	}

	cb.Spec = &apiextensionsv1.JSON{Raw: mustMarshalJSON(s)}
}

func (cb *Dataplane) GetStatus() (core_model.ResourceStatus, error) {
	return nil, nil
}

func (cb *Dataplane) SetStatus(_ core_model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (cb *Dataplane) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *DataplaneList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.Dataplane{}, &Dataplane{
		APIVersion: GroupVersion.String(),
		Kind:       "Dataplane",
	})
	registry.RegisterListType(&mesh_proto.Dataplane{}, &DataplaneList{
		APIVersion: GroupVersion.String(),
		Kind:       "DataplaneList",
	})
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=kuma,scope=Namespaced
type DataplaneInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Mesh is the name of the Kuma mesh this resource belongs to.
	// It may be omitted for cluster-scoped resources.
	//
	// +kubebuilder:validation:Optional
	Mesh string `json:"mesh,omitempty"`
	// Status is the status the Kuma resource.
	// +kubebuilder:validation:Optional
	Status *apiextensionsv1.JSON `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type DataplaneInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataplaneInsight `json:"items"`
}

func init() {
	knownTypes = append(knownTypes, &DataplaneInsight{}, &DataplaneInsightList{})
}

func (cb *DataplaneInsight) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *DataplaneInsight) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *DataplaneInsight) GetMesh() string {
	return cb.Mesh
}

func (cb *DataplaneInsight) SetMesh(mesh string) {
	cb.Mesh = mesh
}

func (cb *DataplaneInsight) GetSpec() (core_model.ResourceSpec, error) {
	spec := cb.Status
	m := mesh_proto.DataplaneInsight{}

	if spec == nil || len(spec.Raw) == 0 {
		return &m, nil
	}

	err := json.Unmarshal(spec.Raw, &m)
	return &m, err
}

func (cb *DataplaneInsight) SetSpec(spec core_model.ResourceSpec) {
	if spec == nil {
		cb.Status = nil
		return
	}

	s, ok := spec.(*mesh_proto.DataplaneInsight)
	if !ok {
		panic(fmt.Sprintf("unexpected spec type %T", spec))
	}

	cb.Status = &apiextensionsv1.JSON{Raw: mustMarshalJSON(s)}
}

func (cb *DataplaneInsight) GetStatus() (core_model.ResourceStatus, error) {
	return nil, nil
}

func (cb *DataplaneInsight) SetStatus(_ core_model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (cb *DataplaneInsight) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *DataplaneInsightList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.DataplaneInsight{}, &DataplaneInsight{
		APIVersion: GroupVersion.String(),
		Kind:       "DataplaneInsight",
	})
	registry.RegisterListType(&mesh_proto.DataplaneInsight{}, &DataplaneInsightList{
		APIVersion: GroupVersion.String(),
		Kind:       "DataplaneInsightList",
	})
}
