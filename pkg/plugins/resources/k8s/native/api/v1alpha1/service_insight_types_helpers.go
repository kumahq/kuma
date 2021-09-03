package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (in *ServiceInsight) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *ServiceInsight) SetObjectMeta(m *metav1.ObjectMeta) {
	in.ObjectMeta = *m
}

func (in *ServiceInsight) GetMesh() string {
	return in.Mesh
}

func (in *ServiceInsight) SetMesh(mesh string) {
	in.Mesh = mesh
}

func (in *ServiceInsight) GetSpec() map[string]interface{} {
	return in.Spec
}

func (in *ServiceInsight) SetSpec(spec map[string]interface{}) {
	in.Spec = spec
}

func (in *ServiceInsight) Scope() model.Scope {
	return model.ScopeCluster
}

func (in *ServiceInsightList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(in.Items))
	for i := range in.Items {
		result[i] = &in.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.ServiceInsight{}, &ServiceInsight{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ServiceInsight",
		},
	})
	registry.RegisterListType(&mesh_proto.ServiceInsight{}, &ServiceInsightList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ServiceInsightList",
		},
	})
}
