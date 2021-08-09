package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (in *MeshInsight) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *MeshInsight) SetObjectMeta(m *metav1.ObjectMeta) {
	in.ObjectMeta = *m
}

func (in *MeshInsight) GetMesh() string {
	return in.Mesh
}

func (in *MeshInsight) SetMesh(mesh string) {
	in.Name = mesh
}

func (in *MeshInsight) GetSpec() map[string]interface{} {
	return in.Spec
}

func (in *MeshInsight) SetSpec(spec map[string]interface{}) {
	in.Spec = spec
}

func (in *MeshInsight) Scope() model.Scope {
	return model.ScopeCluster
}

func (in *MeshInsightList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(in.Items))
	for i := range in.Items {
		result[i] = &in.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.MeshInsight{}, &MeshInsight{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshInsight",
		},
	})
	registry.RegisterListType(&mesh_proto.MeshInsight{}, &MeshInsightList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshInsightList",
		},
	})
}
