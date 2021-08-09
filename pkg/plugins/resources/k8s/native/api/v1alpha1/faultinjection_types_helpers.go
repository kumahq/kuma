package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (fi *FaultInjection) GetObjectMeta() *metav1.ObjectMeta {
	return &fi.ObjectMeta
}

func (fi *FaultInjection) SetObjectMeta(m *metav1.ObjectMeta) {
	fi.ObjectMeta = *m
}

func (fi *FaultInjection) GetMesh() string {
	return fi.Mesh
}

func (fi *FaultInjection) SetMesh(mesh string) {
	fi.Mesh = mesh
}

func (fi *FaultInjection) GetSpec() map[string]interface{} {
	return fi.Spec
}

func (fi *FaultInjection) SetSpec(spec map[string]interface{}) {
	fi.Spec = spec
}

func (fi *FaultInjection) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *FaultInjectionList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.FaultInjection{}, &FaultInjection{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "FaultInjection",
		},
	})
	registry.RegisterListType(&mesh_proto.FaultInjection{}, &FaultInjectionList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "FaultInjectionList",
		},
	})
}
