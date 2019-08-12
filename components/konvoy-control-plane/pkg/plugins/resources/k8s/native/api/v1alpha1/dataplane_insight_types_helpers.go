package v1alpha1

import (
	proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/pkg/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (pt *DataplaneInsight) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *DataplaneInsight) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *DataplaneInsight) GetMesh() string {
	return pt.Mesh
}

func (pt *DataplaneInsight) SetMesh(mesh string) {
	pt.Mesh = mesh
}

func (pt *DataplaneInsight) GetSpec() map[string]interface{} {
	return pt.Status
}

func (pt *DataplaneInsight) SetSpec(spec map[string]interface{}) {
	pt.Status = spec
}

func (l *DataplaneInsightList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&proto.DataplaneInsight{}, &DataplaneInsight{})
	registry.RegisterListType(&proto.DataplaneInsight{}, &DataplaneInsightList{})
}
