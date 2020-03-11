package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (o *HealthCheck) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

func (o *HealthCheck) SetObjectMeta(m *metav1.ObjectMeta) {
	o.ObjectMeta = *m
}

func (o *HealthCheck) GetMesh() string {
	return o.Mesh
}

func (o *HealthCheck) SetMesh(mesh string) {
	o.Mesh = mesh
}

func (o *HealthCheck) GetSpec() map[string]interface{} {
	return o.Spec
}

func (o *HealthCheck) SetSpec(spec map[string]interface{}) {
	o.Spec = spec
}

func (o *HealthCheck) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *HealthCheckList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&proto.HealthCheck{}, &HealthCheck{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "HealthCheck",
		},
	})
	registry.RegisterListType(&proto.HealthCheck{}, &HealthCheckList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "HealthCheckList",
		},
	})
}
