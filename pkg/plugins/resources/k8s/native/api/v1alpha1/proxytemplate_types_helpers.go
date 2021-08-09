package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
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
	return model.ScopeCluster
}

func (l *ProxyTemplateList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.ProxyTemplate{}, &ProxyTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ProxyTemplate",
		},
	})
	registry.RegisterListType(&mesh_proto.ProxyTemplate{}, &ProxyTemplateList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ProxyTemplateList",
		},
	})
}
