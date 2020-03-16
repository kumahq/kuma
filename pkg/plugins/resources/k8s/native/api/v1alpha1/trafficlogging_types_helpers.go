package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (tp *TrafficLog) GetObjectMeta() *metav1.ObjectMeta {
	return &tp.ObjectMeta
}

func (tp *TrafficLog) SetObjectMeta(m *metav1.ObjectMeta) {
	tp.ObjectMeta = *m
}

func (tp *TrafficLog) GetMesh() string {
	return tp.Mesh
}

func (tp *TrafficLog) SetMesh(mesh string) {
	tp.Mesh = mesh
}

func (tp *TrafficLog) GetSpec() map[string]interface{} {
	return tp.Spec
}

func (tp *TrafficLog) SetSpec(spec map[string]interface{}) {
	tp.Spec = spec
}

func (tp *TrafficLog) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *TrafficLogList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&proto.TrafficLog{}, &TrafficLog{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "TrafficLog",
		},
	})
	registry.RegisterListType(&proto.TrafficLog{}, &TrafficLogList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "TrafficLogList",
		},
	})
}
