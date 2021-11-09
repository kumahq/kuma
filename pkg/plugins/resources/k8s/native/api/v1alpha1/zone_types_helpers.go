package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (cb *Zone) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *Zone) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *Zone) GetMesh() string {
	return cb.Mesh
}

func (cb *Zone) SetMesh(mesh string) {
	cb.Mesh = mesh
}

func (cb *Zone) GetSpec() map[string]interface{} {
	return cb.Spec
}

func (cb *Zone) SetSpec(spec map[string]interface{}) {
	cb.Spec = spec
}

func (cb *Zone) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *ZoneList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&system_proto.Zone{}, &Zone{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "Zone",
		},
	})
	registry.RegisterListType(&system_proto.Zone{}, &ZoneList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "ZoneList",
		},
	})
}
