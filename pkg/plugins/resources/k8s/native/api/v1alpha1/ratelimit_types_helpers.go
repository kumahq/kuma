package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (tp *RateLimit) GetObjectMeta() *metav1.ObjectMeta {
	return &tp.ObjectMeta
}

func (tp *RateLimit) SetObjectMeta(m *metav1.ObjectMeta) {
	tp.ObjectMeta = *m
}

func (tp *RateLimit) GetMesh() string {
	return tp.Mesh
}

func (tp *RateLimit) SetMesh(mesh string) {
	tp.Mesh = mesh
}

func (tp *RateLimit) GetSpec() map[string]interface{} {
	return tp.Spec
}

func (tp *RateLimit) SetSpec(spec map[string]interface{}) {
	tp.Spec = spec
}

func (tp *RateLimit) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *RateLimitList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&mesh_proto.RateLimit{}, &RateLimit{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "RateLimit",
		},
	})
	registry.RegisterListType(&mesh_proto.RateLimit{}, &RateLimitList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "RateLimitList",
		},
	})
}
