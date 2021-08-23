package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func (tt *VirtualOutbound) GetObjectMeta() *metav1.ObjectMeta {
	return &tt.ObjectMeta
}

func (tt *VirtualOutbound) SetObjectMeta(m *metav1.ObjectMeta) {
	tt.ObjectMeta = *m
}

func (tt *VirtualOutbound) GetMesh() string {
	return tt.Mesh
}

func (tt *VirtualOutbound) SetMesh(mesh string) {
	tt.Mesh = mesh
}

func (tt *VirtualOutbound) GetSpec() map[string]interface{} {
	return tt.Spec
}

func (tt *VirtualOutbound) SetSpec(spec map[string]interface{}) {
	tt.Spec = spec
}

func (tt *VirtualOutbound) Scope() model.Scope {
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
	registry.RegisterObjectType(&mesh_proto.VirtualOutbound{}, &VirtualOutbound{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "VirtualOutbound",
		},
	})
	registry.RegisterListType(&mesh_proto.VirtualOutbound{}, &VirtualOutboundList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "VirtualOutboundList",
		},
	})
}
