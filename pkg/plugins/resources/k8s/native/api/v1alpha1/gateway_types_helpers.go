package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

func (pt *Gateway) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *Gateway) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *Gateway) GetMesh() string {
	return pt.Mesh
}

func (pt *Gateway) SetMesh(mesh string) {
	pt.Mesh = mesh
}

func (pt *Gateway) GetSpec() map[string]interface{} {
	return pt.Spec
}

func (pt *Gateway) SetSpec(spec map[string]interface{}) {
	pt.Spec = spec
}

func (pt *Gateway) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *GatewayList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}
