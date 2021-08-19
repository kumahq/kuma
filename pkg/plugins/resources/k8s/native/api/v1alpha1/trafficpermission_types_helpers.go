package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (tp *TrafficPermission) GetObjectMeta() *metav1.ObjectMeta {
	return &tp.ObjectMeta
}

func (tp *TrafficPermission) SetObjectMeta(m *metav1.ObjectMeta) {
	tp.ObjectMeta = *m
}

func (tp *TrafficPermission) GetMesh() string {
	return tp.Mesh
}

func (tp *TrafficPermission) SetMesh(mesh string) {
	tp.Mesh = mesh
}

func (tp *TrafficPermission) GetSpec() map[string]interface{} {
	return tp.Spec
}

func (tp *TrafficPermission) SetSpec(spec map[string]interface{}) {
	tp.Spec = spec
}

func (tp *TrafficPermission) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *TrafficPermissionList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.TrafficPermission{}, &TrafficPermission{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "TrafficPermission",
		},
	})
	registry.RegisterListType(&mesh_proto.TrafficPermission{}, &TrafficPermissionList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "TrafficPermissionList",
		},
	})
}
