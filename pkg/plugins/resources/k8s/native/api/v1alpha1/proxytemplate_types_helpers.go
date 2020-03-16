package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (pt *ProxyTemplate) GetObjectMeta() *metav1.ObjectMeta {
	return &pt.ObjectMeta
}

func (pt *ProxyTemplate) SetObjectMeta(m *metav1.ObjectMeta) {
	pt.ObjectMeta = *m
}

func (pt *ProxyTemplate) GetMesh() string {
	return pt.Mesh
}

func (pt *ProxyTemplate) SetMesh(mesh string) {
	pt.Mesh = mesh
}

func (pt *ProxyTemplate) GetSpec() map[string]interface{} {
	return pt.Spec
}

func (pt *ProxyTemplate) SetSpec(spec map[string]interface{}) {
	pt.Spec = spec
}

func (l *ProxyTemplate) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *ProxyTemplateList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&proto.ProxyTemplate{}, &ProxyTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ProxyTemplate",
		},
	})
	registry.RegisterListType(&proto.ProxyTemplate{}, &ProxyTemplateList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ProxyTemplateList",
		},
	})
}
