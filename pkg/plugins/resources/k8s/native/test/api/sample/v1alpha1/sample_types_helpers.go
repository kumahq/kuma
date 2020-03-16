package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

func (pt *SampleTrafficRoute) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *SampleTrafficRoute) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *SampleTrafficRoute) GetMesh() string {
	return pt.Mesh
}

func (pt *SampleTrafficRoute) SetMesh(mesh string) {
	pt.Mesh = mesh
}

func (pt *SampleTrafficRoute) GetSpec() map[string]interface{} {
	return pt.Spec
}

func (pt *SampleTrafficRoute) SetSpec(spec map[string]interface{}) {
	pt.Spec = spec
}

func (pt *SampleTrafficRoute) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *SampleTrafficRouteList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}
