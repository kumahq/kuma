package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (vo *VirtualOutbound) GetObjectMeta() *metav1.ObjectMeta {
	return &vo.ObjectMeta
}

func (vo *VirtualOutbound) SetObjectMeta(m *metav1.ObjectMeta) {
	vo.ObjectMeta = *m
}

func (vo *VirtualOutbound) GetMesh() string {
	return vo.Mesh
}

func (vo *VirtualOutbound) SetMesh(mesh string) {
	vo.Mesh = mesh
}

func (vo *VirtualOutbound) GetSpec() map[string]interface{} {
	return vo.Spec
}

func (vo *VirtualOutbound) SetSpec(spec map[string]interface{}) {
	vo.Spec = spec
}

func (vo *VirtualOutbound) Scope() model.Scope {
	return model.ScopeCluster
}

func (l *VirtualOutboundList) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

func init() {
	registry.RegisterObjectType(&proto.VirtualOutbound{}, &VirtualOutbound{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "VirtualOutbound",
		},
	})
	registry.RegisterListType(&proto.VirtualOutbound{}, &VirtualOutboundList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "VirtualOutboundList",
		},
	})
}
