package v1alpha1

import (
	proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (tp *TrafficLogging) GetObjectMeta() *metav1.ObjectMeta {
	return &tp.ObjectMeta
}

func (tp *TrafficLogging) SetObjectMeta(m *metav1.ObjectMeta) {
	tp.ObjectMeta = *m
}

func (tp *TrafficLogging) GetMesh() string {
	return tp.Mesh
}

func (tp *TrafficLogging) SetMesh(mesh string) {
	tp.Mesh = mesh
}

func (tp *TrafficLogging) GetSpec() map[string]interface{} {
	return tp.Spec
}

func (tp *TrafficLogging) SetSpec(spec map[string]interface{}) {
	tp.Spec = spec
}

func (l *TrafficLoggingList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&proto.TrafficLogging{}, &TrafficLogging{})
	registry.RegisterListType(&proto.TrafficLogging{}, &TrafficLoggingList{})
}
