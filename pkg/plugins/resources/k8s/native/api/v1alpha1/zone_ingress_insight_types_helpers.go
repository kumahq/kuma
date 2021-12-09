package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (cb *ZoneIngressInsight) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *ZoneIngressInsight) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *ZoneIngressInsight) GetMesh() string {
	return cb.Mesh
}

func (cb *ZoneIngressInsight) SetMesh(mesh string) {
	cb.Mesh = mesh
}

func (cb *ZoneIngressInsight) GetSpec() map[string]interface{} {
	return cb.Spec
}

func (cb *ZoneIngressInsight) SetSpec(spec map[string]interface{}) {
	cb.Spec = spec
}

func (cb *ZoneIngressInsight) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *ZoneIngressInsightList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.ZoneIngressInsight{}, &ZoneIngressInsight{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ZoneIngressInsight",
		},
	})
	registry.RegisterListType(&mesh_proto.ZoneIngressInsight{}, &ZoneIngressInsightList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ZoneIngressInsightList",
		},
	})
}
