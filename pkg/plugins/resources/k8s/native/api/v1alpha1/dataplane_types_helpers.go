package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (pt *Dataplane) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *Dataplane) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *Dataplane) GetMesh() string {
	return pt.Mesh
}

func (pt *Dataplane) SetMesh(mesh string) {
	pt.Mesh = mesh
}

func (pt *Dataplane) GetSpec() map[string]interface{} {
	return pt.Spec
}

func (pt *Dataplane) SetSpec(spec map[string]interface{}) {
	pt.Spec = spec
}

func (l *DataplaneList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func (l *Dataplane) Scope() model.Scope {
	return model.ScopeNamespace
}

func init() {
	registry.RegisterObjectType(&proto.Dataplane{}, &Dataplane{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "Dataplane",
		},
	})
	registry.RegisterListType(&proto.Dataplane{}, &DataplaneList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "DataplaneList",
		},
	})
}
