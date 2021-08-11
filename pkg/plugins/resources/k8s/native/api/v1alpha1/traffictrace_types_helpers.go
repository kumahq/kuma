package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (tt *TrafficTrace) GetObjectMeta() *metav1.ObjectMeta {
	return &tt.ObjectMeta
}

func (tt *TrafficTrace) SetObjectMeta(m *metav1.ObjectMeta) {
	tt.ObjectMeta = *m
}

func (tt *TrafficTrace) GetMesh() string {
	return tt.Mesh
}

func (tt *TrafficTrace) SetMesh(mesh string) {
	tt.Mesh = mesh
}

func (tt *TrafficTrace) GetSpec() map[string]interface{} {
	return tt.Spec
}

func (tt *TrafficTrace) SetSpec(spec map[string]interface{}) {
	tt.Spec = spec
}

func (tt *TrafficTrace) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *TrafficTraceList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.TrafficTrace{}, &TrafficTrace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "TrafficTrace",
		},
	})
	registry.RegisterListType(&mesh_proto.TrafficTrace{}, &TrafficTraceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "TrafficTraceList",
		},
	})
}
