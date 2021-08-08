package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (pt *ExternalService) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *ExternalService) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *ExternalService) GetMesh() string {
	return pt.Mesh
}

func (pt *ExternalService) SetMesh(mesh string) {
	pt.Mesh = mesh
}

func (pt *ExternalService) GetSpec() map[string]interface{} {
	return pt.Spec
}

func (pt *ExternalService) SetSpec(spec map[string]interface{}) {
	pt.Spec = spec
}

func (l *ExternalServiceList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func (l *ExternalService) Scope() model.Scope {
	return model.ScopeCluster
}

func init() {
	registry.RegisterObjectType(&mesh_proto.ExternalService{}, &ExternalService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ExternalService",
		},
	})
	registry.RegisterListType(&mesh_proto.ExternalService{}, &ExternalServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ExternalServiceList",
		},
	})
}
