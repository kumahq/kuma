package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (pt *Mesh) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *Mesh) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *Mesh) GetMesh() string {
	return ""
}

func (pt *Mesh) SetMesh(mesh string) {
	pt.Name = mesh
}

func (pt *Mesh) GetSpec() map[string]interface{} {
	return pt.Spec
}

func (pt *Mesh) SetSpec(spec map[string]interface{}) {
	pt.Spec = spec
}

func (pt *Mesh) Scope() model.Scope {
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
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "Mesh",
		},
	})
	registry.RegisterListType(&mesh_proto.Mesh{}, &MeshList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshList",
		},
	})
}
