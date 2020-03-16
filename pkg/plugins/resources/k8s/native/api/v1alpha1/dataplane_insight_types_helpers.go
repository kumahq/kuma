package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
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

func (pt *DataplaneInsight) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *DataplaneInsightList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&proto.DataplaneInsight{}, &DataplaneInsight{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "DataplaneInsight",
		},
	})
	registry.RegisterListType(&proto.DataplaneInsight{}, &DataplaneInsightList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "DataplaneInsightList",
		},
	})
}
