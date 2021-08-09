package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (pt *TrafficRoute) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *TrafficRoute) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *TrafficRoute) GetMesh() string {
	return pt.Mesh
}

func (pt *TrafficRoute) SetMesh(mesh string) {
	pt.Mesh = mesh
}

func (pt *TrafficRoute) GetSpec() map[string]interface{} {
	return pt.Spec
}

func (pt *TrafficRoute) SetSpec(spec map[string]interface{}) {
	pt.Spec = spec
}

func (pt *TrafficRoute) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *TrafficRouteList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.TrafficRoute{}, &TrafficRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "TrafficRoute",
		},
	})
	registry.RegisterListType(&mesh_proto.TrafficRoute{}, &TrafficRouteList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "TrafficRouteList",
		},
	})
}
