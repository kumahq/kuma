package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (cb *ZoneInsight) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *ZoneInsight) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *ZoneInsight) GetMesh() string {
	return cb.Mesh
}

func (cb *ZoneInsight) SetMesh(mesh string) {
	cb.Mesh = mesh
}

func (cb *ZoneInsight) GetSpec() map[string]interface{} {
	return cb.Spec
}

func (cb *ZoneInsight) SetSpec(spec map[string]interface{}) {
	cb.Spec = spec
}

func (cb *ZoneInsight) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *ZoneInsightList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&system_proto.ZoneInsight{}, &ZoneInsight{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ZoneInsight",
		},
	})
	registry.RegisterListType(&system_proto.ZoneInsight{}, &ZoneInsightList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ZoneInsightList",
		},
	})
}
