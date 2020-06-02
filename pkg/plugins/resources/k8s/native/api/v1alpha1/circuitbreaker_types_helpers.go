package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (cb *CircuitBreaker) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *CircuitBreaker) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *CircuitBreaker) GetMesh() string {
	return cb.Mesh
}

func (cb *CircuitBreaker) SetMesh(mesh string) {
	cb.Mesh = mesh
}

func (cb *CircuitBreaker) GetSpec() map[string]interface{} {
	return cb.Spec
}

func (cb *CircuitBreaker) SetSpec(spec map[string]interface{}) {
	cb.Spec = spec
}

func (cb *CircuitBreaker) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *CircuitBreakerList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&proto.CircuitBreaker{}, &CircuitBreaker{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "CircuitBreaker",
		},
	})
	registry.RegisterListType(&proto.CircuitBreaker{}, &CircuitBreakerList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "CircuitBreakerList",
		},
	})
}
