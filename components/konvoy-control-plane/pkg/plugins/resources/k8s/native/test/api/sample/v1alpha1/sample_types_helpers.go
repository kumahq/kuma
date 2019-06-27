package v1alpha1

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (pt *TrafficRoute) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *TrafficRoute) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *TrafficRoute) GetSpec() map[string]interface{} {
	return pt.Spec
}

func (pt *TrafficRoute) SetSpec(spec map[string]interface{}) {
	pt.Spec = spec
}

func (l *TrafficRouteList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}
