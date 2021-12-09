package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (o *Retry) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

func (o *Retry) SetObjectMeta(m *metav1.ObjectMeta) {
	o.ObjectMeta = *m
}

func (o *Retry) GetMesh() string {
	return o.Mesh
}

func (o *Retry) SetMesh(mesh string) {
	o.Mesh = mesh
}

func (o *Retry) GetSpec() map[string]interface{} {
	return o.Spec
}

func (o *Retry) SetSpec(spec map[string]interface{}) {
	o.Spec = spec
}

func (o *Retry) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *RetryList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.Retry{}, &Retry{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "Retry",
		},
	})
	registry.RegisterListType(&mesh_proto.Retry{}, &RetryList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "RetryList",
		},
	})
}
