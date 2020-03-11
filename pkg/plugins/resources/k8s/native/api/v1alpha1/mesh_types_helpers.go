package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (pt *Mesh) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *Mesh) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *Mesh) GetMesh() string {
	return pt.Name
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
	registry.RegisterObjectType(&proto.Mesh{}, &Mesh{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "Mesh",
		},
	})
	registry.RegisterListType(&proto.Mesh{}, &MeshList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshList",
		},
	})
}
