package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (t *Timeout) GetObjectMeta() *metav1.ObjectMeta {
	return &t.ObjectMeta
}

func (t *Timeout) SetObjectMeta(m *metav1.ObjectMeta) {
	t.ObjectMeta = *m
}

func (t *Timeout) GetMesh() string {
	return t.Mesh
}

func (t *Timeout) SetMesh(mesh string) {
	t.Mesh = mesh
}

func (t *Timeout) GetSpec() map[string]interface{} {
	return t.Spec
}

func (t *Timeout) SetSpec(spec map[string]interface{}) {
	t.Spec = spec
}

func (t *Timeout) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *TimeoutList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.Timeout{}, &Timeout{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "Timeout",
		},
	})
	registry.RegisterListType(&mesh_proto.Timeout{}, &TimeoutList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "TimeoutList",
		},
	})
}
