package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (zi *ZoneIngress) GetObjectMeta() *metav1.ObjectMeta {
	return &zi.ObjectMeta
}

func (zi *ZoneIngress) SetObjectMeta(m *metav1.ObjectMeta) {
	zi.ObjectMeta = *m
}

func (zi *ZoneIngress) GetMesh() string {
	return zi.Mesh
}

func (zi *ZoneIngress) SetMesh(mesh string) {
	zi.Mesh = mesh
}

func (zi *ZoneIngress) GetSpec() map[string]interface{} {
	return zi.Spec
}

func (zi *ZoneIngress) SetSpec(spec map[string]interface{}) {
	zi.Spec = spec
}

func (zi *ZoneIngress) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *ZoneIngressList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&proto.ZoneIngress{}, &ZoneIngress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ZoneIngress",
		},
	})
	registry.RegisterListType(&proto.ZoneIngress{}, &ZoneIngressList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ZoneIngressList",
		},
	})
}
